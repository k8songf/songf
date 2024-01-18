package controller

import (
	"fmt"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"sync"
	alpha1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

type jobItemTree struct {
	sync.RWMutex

	name      string
	uuid      types.UID
	nameSpace string

	startItemNode *itemNode

	workNodes map[string]*itemNode

	itemStatus map[string]*v1alpha1.ItemStatus
}

// todo 先是一个空的，然后build from 组件或者任务
func newJobItemTree(job *v1alpha1.Job) (*jobItemTree, error) {

	nodeMap := map[string]*itemNode{}
	var fatherNode *itemNode
	itemStatus := map[string]*v1alpha1.ItemStatus{}

	for _, item := range job.Spec.Items {
		if _, ok := nodeMap[item.Name]; ok {
			return nil, fmt.Errorf("new job item tree build err: item name %s repeated", item.Name)
		}

		var itemImpl *v1alpha1.Item
		itemImpl = &item
		nodeMap[item.Name] = &itemNode{
			item: itemImpl,
		}

		if len(item.RunAfter) == 0 {
			if fatherNode != nil {
				return nil, fmt.Errorf("new job item tree build err: muilty start item")
			} else {
				fatherNode = nodeMap[item.Name]
			}
		}

		flag := len(job.Status.ItemStatus) == 0
		if _, ok := job.Status.ItemStatus[item.Name]; !ok {
			flag = true
		}

		if flag {
			itemStatus[item.Name] = &v1alpha1.ItemStatus{
				Name:  item.Name,
				Phase: v1alpha1.ItemPending,
			}
		} else {
			status := job.Status.ItemStatus[item.Name]
			itemStatus[item.Name] = &status
		}
	}

	for _, node := range nodeMap {
		var nodeImpl *itemNode
		nodeImpl = node

		for _, parentName := range node.item.RunAfter {
			if _, ok := nodeMap[parentName]; !ok {
				return nil, fmt.Errorf("not find parent item name %s", parentName)
			} else {
				nodeMap[parentName].child = append(nodeMap[parentName].child, nodeImpl)
			}
		}
	}

	tree := &jobItemTree{
		name:      job.Name,
		uuid:      job.UID,
		nameSpace: job.Namespace,

		startItemNode: fatherNode,
		workNodes:     nodeMap,
		itemStatus:    itemStatus,
	}

	if tree.hasCycle() {
		return nil, fmt.Errorf("new job item tree build err: items tree has cycle")
	}

	return tree, nil
}

func (t *jobItemTree) hasCycle() bool {
	return t.hasCycleDfs(t.startItemNode, map[string]bool{})
}

func (t *jobItemTree) hasCycleDfs(node *itemNode, visited map[string]bool) bool {

	if visited[node.item.Name] {
		return true
	}

	visited[node.item.Name] = true

	for _, child := range node.child {
		if t.hasCycleDfs(child, visited) {
			return true
		}
	}

	visited[node.item.Name] = false

	return false
}

func (t *jobItemTree) syncFromKubeJob(job *v1.Job) error {

	fn := func(status *v1alpha1.ItemStatus) {
		jobState, ok := status.JobStatus[job.Name]
		if !ok {
			jobState = alpha1.JobState{
				Phase:              alpha1.Pending,
				LastTransitionTime: job.CreationTimestamp,
			}
		}

		if job.DeletionTimestamp != nil && !job.DeletionTimestamp.IsZero() {
			jobState = alpha1.JobState{
				Phase:              alpha1.Terminated,
				LastTransitionTime: *job.DeletionTimestamp,
			}
			status.JobStatus[job.Name] = jobState
			return
		}

		conditionsLength := len(job.Status.Conditions)
		if conditionsLength > 0 {
			condition := job.Status.Conditions[conditionsLength-1]

			jobState.LastTransitionTime = condition.LastTransitionTime
			jobState.Message = condition.Message
			jobState.Reason = condition.Reason

			switch condition.Type {
			case v1.JobSuspended:
				jobState.Phase = alpha1.Aborted

			case v1.JobComplete:
				jobState.Phase = alpha1.Completed

			case v1.JobFailureTarget, v1.JobFailed:
				jobState.Phase = alpha1.Failed

			default:
				klog.Errorf("can not recognize kube job %s/%s condition type: %s", job.Namespace, job.Name, condition.Type)
			}
		}

		status.JobStatus[job.Name] = jobState
		return

	}

	return t.syncFromObject(job, fn)

}

func (t *jobItemTree) syncFromVcJob(job *alpha1.Job) error {

	fn := func(status *v1alpha1.ItemStatus) {
		status.JobStatus[job.Name] = job.Status.State
	}

	return t.syncFromObject(job, fn)
}

func (t *jobItemTree) syncFromService(service *corev1.Service) error {

	fn := func(status *v1alpha1.ItemStatus) {
		serviceStatus, ok := status.ServiceStatus[service.Name]
		if !ok {
			serviceStatus = v1alpha1.RegularModuleStatus{
				Phase: v1alpha1.RegularModuleUnknown,
			}
		}

		if service.DeletionTimestamp == nil || service.DeletionTimestamp.IsZero() {

			serviceStatus.Phase = v1alpha1.RegularModuleCreated
			serviceStatus.LastTransitionTime = service.CreationTimestamp
		} else {
			serviceStatus.Phase = v1alpha1.RegularModuleFailed
			serviceStatus.LastTransitionTime = *service.DeletionTimestamp
		}
		status.ServiceStatus[service.Name] = serviceStatus
	}

	return t.syncFromObject(service, fn)
}

func (t *jobItemTree) syncFromConfigmap(configmap *corev1.ConfigMap) error {

	fn := func(status *v1alpha1.ItemStatus) {
		cmStatus, ok := status.ConfigMapStatus[configmap.Name]
		if !ok {
			cmStatus = v1alpha1.RegularModuleStatus{
				Phase: v1alpha1.RegularModuleUnknown,
			}
		}

		if configmap.DeletionTimestamp == nil || configmap.DeletionTimestamp.IsZero() {

			cmStatus.Phase = v1alpha1.RegularModuleCreated
			cmStatus.LastTransitionTime = configmap.CreationTimestamp
		} else {
			cmStatus.Phase = v1alpha1.RegularModuleFailed
			cmStatus.LastTransitionTime = *configmap.DeletionTimestamp
		}
		status.ServiceStatus[configmap.Name] = cmStatus
	}

	return t.syncFromObject(configmap, fn)
}

func (t *jobItemTree) syncFromSecret(secret *corev1.Secret) error {

	fn := func(status *v1alpha1.ItemStatus) {
		secretStatus, ok := status.SecretStatus[secret.Name]
		if !ok {
			secretStatus = v1alpha1.RegularModuleStatus{
				Phase: v1alpha1.RegularModuleUnknown,
			}
		}

		if secret.DeletionTimestamp == nil || secret.DeletionTimestamp.IsZero() {

			secretStatus.Phase = v1alpha1.RegularModuleCreated
			secretStatus.LastTransitionTime = secret.CreationTimestamp
		} else {
			secretStatus.Phase = v1alpha1.RegularModuleFailed
			secretStatus.LastTransitionTime = *secret.DeletionTimestamp
		}
		status.SecretStatus[secret.Name] = secretStatus
	}

	return t.syncFromObject(secret, fn)
}

func (t *jobItemTree) syncFromObject(object client.Object, fn func(status *v1alpha1.ItemStatus)) error {
	_, itemName := getJobNameAndItemNameFromObject(object)

	t.Lock()
	defer t.Unlock()

	status, ok := t.itemStatus[itemName]
	if !ok {
		return fmt.Errorf("not found item %s from %s/%s tree", itemName, t.nameSpace, t.name)
	}

	switch status.Phase {
	case v1alpha1.ItemScheduling, v1alpha1.ItemScheduled:
		fn(status)

	case v1alpha1.ItemPending:

		return fmt.Errorf("received created %s %s but item %s/%s is pending", object.GetObjectKind(), object.GetName(), t.nameSpace, itemName)

	default:

		klog.Warningf("received created %s %s and item %s/%s is %s", object.GetObjectKind(), object.GetName(), t.nameSpace, itemName, status.Phase)
	}

	node, ok := t.workNodes[itemName]
	if !ok {
		return fmt.Errorf("can not find item %s from cache tree %s/%s", itemName, t.nameSpace, t.name)
	}

	phase := v1alpha1.ItemCompleted

	jobStateNotFoundNum := 0

	if status.RunningJobNum != nil {
		*status.RunningJobNum = 0
	}
	if status.CompletedJobNum != nil {
		*status.CompletedJobNum = 0
	}
	if status.FailedJobNum != nil {
		*status.FailedJobNum = 0
	}

	for _, job := range node.item.ItemJobs.Jobs {
		name := calItemSubName(t.name, itemName, job.Name)

		state, ok := status.JobStatus[name]
		if !ok {
			jobStateNotFoundNum++
			continue
		}

		switch state.Phase {
		case alpha1.Running:
			if status.RunningJobNum == nil {
				var i int32
				status.RunningJobNum = &i
			}
			*status.RunningJobNum++

		case alpha1.Completed, alpha1.Completing:
			if status.CompletedJobNum == nil {
				var i int32
				status.CompletedJobNum = &i
			}
			*status.CompletedJobNum++

		case alpha1.Failed:
			if status.FailedJobNum == nil {
				var i int32
				status.FailedJobNum = &i
			}
			*status.FailedJobNum++
		}
	}

	if status.FailedJobNum != nil && *status.FailedJobNum > 0 {
		phase = v1alpha1.ItemFailed
	} else if status.CompletedJobNum != nil && *status.CompletedJobNum == int32(len(node.item.ItemJobs.Jobs)) {
		phase = v1alpha1.ItemCompleted
	} else {
		phase = v1alpha1.ItemScheduled
	}

	if phase == v1alpha1.ItemCompleted {
		for _, svc := range node.item.ItemModules.Services {
			svcName := calItemSubName(t.name, itemName, svc.Name)
			state, ok := status.ServiceStatus[svcName]
			if !ok {
				phase = v1alpha1.ItemScheduled
				break
			}

			switch state.Phase {
			case v1alpha1.RegularModuleFailed:
				phase = v1alpha1.ItemFailed
			case v1alpha1.RegularModuleUnknown, v1alpha1.RegularModuleCreating:
				phase = v1alpha1.ItemScheduled
			}

			if phase != v1alpha1.ItemCompleted {
				break
			}
		}
	}

	if phase == v1alpha1.ItemCompleted {

		for _, cm := range node.item.ItemModules.ConfigMaps {
			name := calItemSubName(t.name, itemName, cm.ConfigMap.Name)
			state, ok := status.ConfigMapStatus[name]
			if !ok {
				phase = v1alpha1.ItemScheduled
				break
			}

			switch state.Phase {
			case v1alpha1.RegularModuleFailed:
				phase = v1alpha1.ItemFailed
			case v1alpha1.RegularModuleUnknown, v1alpha1.RegularModuleCreating:
				phase = v1alpha1.ItemScheduled
			}

			if phase != v1alpha1.ItemCompleted {
				break
			}
		}

	}

	if phase == v1alpha1.ItemCompleted {

		for range node.item.ItemModules.Secrets {

			for _, secret := range node.item.ItemModules.Secrets {
				name := calItemSubName(t.name, itemName, secret.Secret.Name)
				state, ok := status.SecretStatus[name]
				if !ok {
					phase = v1alpha1.ItemScheduled
					break
				}

				switch state.Phase {
				case v1alpha1.RegularModuleFailed:
					phase = v1alpha1.ItemFailed
				case v1alpha1.RegularModuleUnknown, v1alpha1.RegularModuleCreating:
					phase = v1alpha1.ItemScheduled
				}

				if phase != v1alpha1.ItemCompleted {
					break
				}
			}
		}
	}

	status.Phase = phase
	t.itemStatus[itemName] = status

	return nil
}

type itemNode struct {
	item  *v1alpha1.Item
	child []*itemNode
}

func (n *itemNode) children() []*itemNode {
	return n.child
}

func (n *itemNode) isFirst() bool {
	if len(n.item.RunAfter) == 0 {
		return true
	}

	return false
}

func (n *itemNode) isLast() bool {
	if len(n.child) == 0 {
		return true
	}

	return false
}
