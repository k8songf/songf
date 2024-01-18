package controller

import (
	"fmt"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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
	_, itemName := getJobNameAndItemNameFromObject(job)

	t.Lock()
	defer t.Unlock()

	_, ok := t.itemStatus[itemName]
	if !ok {
		return fmt.Errorf("not found item %s from %s/%s tree", itemName, t.nameSpace, t.name)
	}

	return nil

}

func (t *jobItemTree) syncFromVcJob(job *alpha1.Job) error {

	return nil
}

func (t *jobItemTree) syncFromService(service *corev1.Service) error {
	_, itemName := getJobNameAndItemNameFromObject(service)

	t.Lock()
	defer t.Unlock()

	status, ok := t.itemStatus[itemName]
	if !ok {
		return fmt.Errorf("not found item %s from %s/%s tree", itemName, t.nameSpace, t.name)
	}

	serviceStatus, ok := status.ServiceStatus[service.Name]
	if !ok {
		serviceStatus = v1alpha1.RegularModuleStatus{
			Phase: v1alpha1.RegularModuleUnknown,
		}
	}

	switch status.Phase {
	case v1alpha1.ItemScheduling, v1alpha1.ItemScheduled:

		if service.DeletionTimestamp == nil || service.DeletionTimestamp.IsZero() {

			serviceStatus.Phase = v1alpha1.RegularModuleCreated
			serviceStatus.LastTransitionTime = service.CreationTimestamp
			status.ServiceStatus[service.Name] = serviceStatus
		} else {
			serviceStatus.Phase = v1alpha1.RegularModuleFailed
			serviceStatus.LastTransitionTime = *service.DeletionTimestamp
			status.ServiceStatus[service.Name] = serviceStatus
		}

		return nil

	case v1alpha1.ItemPending:

		return fmt.Errorf("received created service %s but item %s/%s is pending", service.Name, t.nameSpace, itemName)

	default:

		klog.Warningf("received created service %s and item %s/%s is %s", service.Name, t.nameSpace, itemName, status.Phase)
		return nil
	}

}

func (t *jobItemTree) syncFromConfigmap(configmap *corev1.ConfigMap) error {

	return nil
}

func (t *jobItemTree) syncFromSecret(secret *corev1.Secret) error {

	return nil
}

func (t *jobItemTree) syncItemStatusPhase(itemName string) error {

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
