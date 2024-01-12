package controller

import (
	"k8s.io/apimachinery/pkg/types"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"sync"
)

var Cache *jobCache

func InitializeCache() {
	Cache = newJobCache()
}

type jobCache struct {
	sync.RWMutex

	jobItemTreeCache map[types.UID]*jobItemTree
}

func newJobCache() *jobCache {
	cache := &jobCache{
		jobItemTreeCache: map[types.UID]*jobItemTree{},
	}

	return cache
}

func (c *jobCache) syncJobTree(job *appsv1alpha1.Job) ([]*itemNode, error) {

	var err error

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[job.UID]
	if !ok {
		tree, err = newJobItemTree(job)
		if err != nil {
			return nil, err
		}

		c.jobItemTreeCache[job.UID] = tree
	}

	// find scheduled item and create resource
	var schedulingItem []*itemNode

	for name, status := range tree.itemStatus {
		if status.Phase == appsv1alpha1.ItemScheduling {
			schedulingItem = append(schedulingItem, tree.workNodes[name])
		}
	}

	return schedulingItem, nil
}
