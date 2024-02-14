package controller

import (
	"fmt"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"songf.sh/songf/pkg/job_graph"
	"sync"
)

type jobCache struct {
	sync.RWMutex

	jobItemGraphCache map[string]*job_graph.JobItemGraph
}

func newJobCache() *jobCache {
	cache := &jobCache{
		jobItemGraphCache: map[string]*job_graph.JobItemGraph{},
	}

	return cache
}

func (c *jobCache) syncGraphFromJob(job *appsv1alpha1.Job) error {

	c.Lock()
	defer c.Unlock()

	switch job.Status.State.Phase {
	case appsv1alpha1.Terminated:
		delete(c.jobItemGraphCache, job.Name)
	default:
		graph, ok := c.jobItemGraphCache[job.Name]
		if !ok {
			graph = job_graph.NewJobItemGraph()
		}
		if err := graph.SyncFromJob(job); err != nil {
			return err
		}
		graph.SyncStatusPhase()
		c.jobItemGraphCache[job.Name] = graph
	}

	return nil
}

func (c *jobCache) syncJobItemStatus(job *appsv1alpha1.Job) (bool, error) {
	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[job.Name]
	if !ok {
		return false, fmt.Errorf("not found job %s from graph", job.Name)
	}

	changedFlag := false

	status := graph.GetAllItemStatus()

	for _, item := range job.Spec.Items {
		if item.Truncated != nil && *item.Truncated == true {
			continue
		}

		cacheStatus, ok := status[item.Name]
		if !ok {
			return false, fmt.Errorf("not found job %s item %s status from graph", job.Name, item.Name)
		}

		itemStatus, ok := job.Status.ItemStatus[item.Name]
		if !ok {
			job.Status.ItemStatus[item.Name] = *cacheStatus.DeepCopy()
			changedFlag = true
			continue
		}

		if !apiequality.Semantic.DeepEqual(cacheStatus, itemStatus) {
			job.Status.ItemStatus[item.Name] = *cacheStatus.DeepCopy()
			changedFlag = true
		}

	}

	return changedFlag, nil

}

func (c *jobCache) getFirstJobItem(jobName string) (*appsv1alpha1.Item, bool) {

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		return nil, false
	}

	node, ok := graph.GetStartItemNode()
	if !ok {
		return nil, false
	}
	return node.Item, true

}

func (c *jobCache) getNextScheduleJobItem(jobName string) ([]*appsv1alpha1.Item, bool) {
	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		return nil, false
	}

	nextItems := graph.ItemsNext2Scheduled()
	if len(nextItems) == 0 {
		return nil, false
	}
	return nextItems, true

}

func (c *jobCache) isJobFinished(jobName string) (finished, failed bool, err error) {
	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		return false, false, fmt.Errorf("not found job %s from graph", jobName)
	}

	allItemStatus := graph.GetAllItemStatus()

	startItemNode, ok := graph.GetStartItemNode()
	if !ok {
		return false, false, fmt.Errorf("not found job start item %s from graph", jobName)
	}

	finished, failed = c.isJobFinishedImpl(allItemStatus, startItemNode)

	return finished, failed, nil

}

func (c *jobCache) isJobFinishedImpl(allStatus map[string]*appsv1alpha1.ItemStatus, node *appsv1alpha1.ItemNode) (finished, failed bool) {
	status := allStatus[node.Item.Name]
	switch status.Phase {
	case appsv1alpha1.ItemFailed:
		return true, true
	case appsv1alpha1.ItemCompleted:
		allSubFinished, hasSubFailed := true, false
		for _, child := range node.Child {
			if child.Item.Truncated != nil && *child.Item.Truncated == true {
				continue
			}

			subFinished, subFailed := c.isJobFinishedImpl(allStatus, child)
			if subFailed {
				hasSubFailed = true
			}
			if !subFinished {
				allSubFinished = false
			}
		}
		if hasSubFailed {
			return true, true
		}
		if allSubFinished {
			return true, false
		} else {
			return false, false
		}

	default:
		for _, child := range node.Child {
			if child.Item.Truncated != nil && *child.Item.Truncated == true {
				continue
			}

			_, subFailed := c.isJobFinishedImpl(allStatus, child)
			if subFailed {
				return true, true
			}
		}

		return false, false

	}

}

func (c *jobCache) isJobDeleted(jobName string) (bool, error) {
	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		return false, fmt.Errorf("not found job %s from graph", jobName)
	}

	if graph.DeleteTimestamp != nil && !graph.DeleteTimestamp.IsZero() {
		return true, nil
	}

	return false, nil

}
