package controller

import (
	"fmt"
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

func newJobItemTree() *jobItemTree {
	return &jobItemTree{
		workNodes:  map[string]*itemNode{},
		itemStatus: map[string]*v1alpha1.ItemStatus{},
	}
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

// =============================================
func (t *jobItemTree) syncFromJob(job *v1alpha1.Job) error {

	t.name = job.Name
	t.uuid = job.UID
	t.nameSpace = job.Namespace

	nodeMap := map[string]*itemNode{}
	var fatherNode *itemNode

	for _, item := range job.Spec.Items {
		if _, ok := nodeMap[item.Name]; ok {
			return fmt.Errorf("new job item tree build err: item name %s repeated", item.Name)
		}

		var itemImpl *v1alpha1.Item
		itemImpl = &item
		nodeMap[item.Name] = &itemNode{
			item: itemImpl,
		}

		if len(item.RunAfter) == 0 {
			if fatherNode != nil {
				return fmt.Errorf("new job item tree build err: muilty start item")
			} else {
				fatherNode = nodeMap[item.Name]
			}
		}

		_, ok := t.itemStatus[item.Name]
		if !ok {
			flag := len(job.Status.ItemStatus) == 0
			if _, ok := job.Status.ItemStatus[item.Name]; !ok {
				flag = true
			}

			if flag {
				t.itemStatus[item.Name] = &v1alpha1.ItemStatus{
					Name:  item.Name,
					Phase: v1alpha1.ItemPending,
				}
			} else {
				status := job.Status.ItemStatus[item.Name]
				t.itemStatus[item.Name] = &status
			}
		}

	}

	for _, node := range nodeMap {
		var nodeImpl *itemNode
		nodeImpl = node

		for _, parentName := range node.item.RunAfter {
			if _, ok := nodeMap[parentName]; !ok {
				return fmt.Errorf("not find parent item name %s", parentName)
			} else {
				nodeMap[parentName].child = append(nodeMap[parentName].child, nodeImpl)
			}
		}
	}
	t.startItemNode = fatherNode
	t.workNodes = nodeMap

	if t.hasCycle() {
		return fmt.Errorf("job %s is not a directed acyclic graph", job.Name)
	}

	return nil
}

// =============================================
func (t *jobItemTree) syncFromObject(object client.Object, fn func(status *v1alpha1.ItemStatus)) error {
	_, itemName := getJobNameAndItemNameFromObject(object)

	t.Lock()
	defer t.Unlock()

	status, ok := t.itemStatus[itemName]
	if !ok {
		klog.Infof("not found item %s from %s/%s tree while sync object, create a scheduling one", itemName, t.nameSpace, t.name)
		status = &v1alpha1.ItemStatus{
			Phase: v1alpha1.ItemScheduling,
			Name:  itemName,
		}
	}

	switch status.Phase {
	case v1alpha1.ItemScheduling, v1alpha1.ItemScheduled:
		fn(status)

	case v1alpha1.ItemPending:

		return fmt.Errorf("received created %s %s but item %s is pending", object.GetObjectKind(), object.GetName(), itemName)

	default:

		klog.Warningf("received created %s %s and item %s is %s", object.GetObjectKind(), object.GetName(), itemName, status.Phase)
	}

	if status.Phase == "" {
		status.Phase = v1alpha1.ItemScheduled
	}

	t.itemStatus[itemName] = status

	t.syncItemStatusPhase(itemName)

	return nil
}

func (t *jobItemTree) syncStatusPhase() {
	for itemName, _ := range t.workNodes {
		t.syncItemStatusPhase(itemName)
	}
}

func (t *jobItemTree) syncItemStatusPhase(itemName string) {
	status, ok := t.itemStatus[itemName]
	if !ok {
		klog.Infof("not found item %s from %s/%s tree", itemName, t.nameSpace, t.name)
		return
	}

	workNode, ok := t.workNodes[itemName]
	if !ok {
		klog.Infof("can not find item %s from cache tree %s/%s", itemName, t.nameSpace, t.name)
		return
	}

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

	for _, job := range workNode.item.ItemJobs.Jobs {
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
		status.Phase = v1alpha1.ItemFailed
	} else if status.CompletedJobNum != nil && *status.CompletedJobNum == int32(len(workNode.item.ItemJobs.Jobs)) {
		status.Phase = v1alpha1.ItemCompleted
	} else if jobStateNotFoundNum == 0 {
		status.Phase = v1alpha1.ItemScheduled
	}

	t.itemStatus[itemName] = status

	return
}

func (t *jobItemTree) itemsNext2Scheduled() []*v1alpha1.Item {
	var res []*v1alpha1.Item

	for itemName, itemStatus := range t.itemStatus {
		if itemStatus.Phase != v1alpha1.ItemPending {
			continue
		}

		flag := true
		for _, fatherItemName := range t.workNodes[itemName].item.RunAfter {
			status, ok := t.itemStatus[fatherItemName]
			if !ok {
				flag = false
				break
			}

			if status.Phase != v1alpha1.ItemCompleted {
				flag = false
				break
			}
		}

		if flag {
			res = append(res, t.workNodes[itemName].item.DeepCopy())
		}
	}

	return res
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
