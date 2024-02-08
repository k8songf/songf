package controller

import (
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"songf.sh/songf/pkg/job_graph"
	"sync"
)

var Cache *jobCache

func InitializeCache() {
	Cache = newJobCache()
}

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

func (c *jobCache) syncJobTree(job *appsv1alpha1.Job) error {

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[job.Name]
	if !ok {
		graph = job_graph.NewJobItemGraph()
	}
	if err := graph.SyncFromJob(job); err != nil {
		return err
	}
	graph.SyncStatusPhase()
	c.jobItemGraphCache[job.Name] = graph

	return nil
}

func (c *jobCache) getFirstJobItem(jobName string) (*appsv1alpha1.Item, bool) {

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		return nil, false
	}

	return graph.GetStartItem()

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
