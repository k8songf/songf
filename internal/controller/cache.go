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

func (c *jobCache) syncJobTree(job *appsv1alpha1.Job) error {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.jobItemTreeCache[job.UID]; ok {
		return nil
	}

	tree, err := newJobItemTree(job)
	if err != nil {
		return err
	}

	c.jobItemTreeCache[job.UID] = tree

	// todo find scheduled item and create resource

	return nil
}
