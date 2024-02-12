package job_graph

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"sync"
	alpha1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

type JobItemGraph struct {
	sync.RWMutex

	Name      string
	Uuid      types.UID
	NameSpace string

	startItemNode *v1alpha1.ItemNode

	workNodes map[string]*v1alpha1.ItemNode

	itemStatus map[string]*v1alpha1.ItemStatus
}

func NewJobItemGraph() *JobItemGraph {
	return &JobItemGraph{
		startItemNode: &v1alpha1.ItemNode{},
		workNodes:     map[string]*v1alpha1.ItemNode{},
		itemStatus:    map[string]*v1alpha1.ItemStatus{},
	}
}

func (t *JobItemGraph) GetStartItem() (*v1alpha1.Item, bool) {
	if t.startItemNode == nil || t.startItemNode.Item == nil {
		return nil, false
	}
	return t.startItemNode.Item, true
}

func (t *JobItemGraph) SyncFromJob(job *v1alpha1.Job) error {

	var err error

	t.Name = job.Name
	t.Uuid = job.UID
	t.NameSpace = job.Namespace

	for _, item := range job.Spec.Items {
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

	t.startItemNode, t.workNodes, err = v1alpha1.NewGraphFromJob(job)
	if err != nil {
		return err
	}

	if !v1alpha1.IsJobHasCycleDfs(t.startItemNode, map[string]bool{}) {
		return fmt.Errorf("job %s is not a directed acyclic graph", job.Name)
	}

	return nil
}

func (t *JobItemGraph) SyncFromObject(object client.Object, fn func(status *v1alpha1.ItemStatus)) error {
	_, itemName := v1alpha1.GetJobNameAndItemNameFromObject(object)

	t.Lock()
	defer t.Unlock()

	status, ok := t.itemStatus[itemName]
	if !ok {
		klog.Infof("not found item %s from %s/%s tree while sync object, create a scheduling one", itemName, t.NameSpace, t.Name)
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

func (t *JobItemGraph) SyncStatusPhase() {
	for itemName, _ := range t.workNodes {
		t.syncItemStatusPhase(itemName)
	}
}

func (t *JobItemGraph) syncItemStatusPhase(itemName string) {
	status, ok := t.itemStatus[itemName]
	if !ok {
		klog.Infof("not found item %s from %s/%s tree", itemName, t.NameSpace, t.Name)
		return
	}

	workNode, ok := t.workNodes[itemName]
	if !ok {
		klog.Infof("can not find item %s from cache tree %s/%s", itemName, t.NameSpace, t.Name)
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

	for _, job := range workNode.Item.ItemJobs.Jobs {
		name := v1alpha1.CalJobItemSubName(t.Name, itemName, job.Name)

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
	} else if status.CompletedJobNum != nil && *status.CompletedJobNum == int32(len(workNode.Item.ItemJobs.Jobs)) {
		status.Phase = v1alpha1.ItemCompleted
	} else if jobStateNotFoundNum == 0 {
		status.Phase = v1alpha1.ItemScheduled
	}

	t.itemStatus[itemName] = status

	return
}

func (t *JobItemGraph) ItemsNext2Scheduled() []*v1alpha1.Item {
	var res []*v1alpha1.Item

	for itemName, itemStatus := range t.itemStatus {
		if itemStatus.Phase != v1alpha1.ItemPending {
			continue
		}

		flag := true
		for _, fatherItemName := range t.workNodes[itemName].Item.RunAfter {
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
			res = append(res, t.workNodes[itemName].Item.DeepCopy())
		}
	}

	return res
}
