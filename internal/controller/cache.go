package controller

import (
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"sync"
)

var Cache *jobCache

func InitializeCache() {
	Cache = newJobCache()
}

type jobCache struct {
	sync.RWMutex

	jobItemTreeCache map[string]*jobItemTree
}

func newJobCache() *jobCache {
	cache := &jobCache{
		jobItemTreeCache: map[string]*jobItemTree{},
	}

	return cache
}

func (c *jobCache) syncJobTree(job *appsv1alpha1.Job) error {

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[job.Name]
	if !ok {
		tree = newJobItemTree()
	}
	if err := tree.syncFromJob(job); err != nil {
		return err
	}
	tree.syncStatusPhase()
	c.jobItemTreeCache[job.Name] = tree

	return nil
}

func (c *jobCache) getFirstJobItem(jobName string) (*appsv1alpha1.Item, bool) {

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		return nil, false
	}

	if tree.startItemNode == nil || tree.startItemNode.item == nil {
		return nil, false
	}

	return tree.startItemNode.item, true

}

func (c *jobCache) getNextScheduleJobItem(jobName string) ([]*appsv1alpha1.Item, bool) {
	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		return nil, false
	}

	nextItems := tree.itemsNext2Scheduled()
	if len(nextItems) == 0 {
		return nil, false
	}
	return nextItems, true

}
