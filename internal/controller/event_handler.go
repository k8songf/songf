package controller

import (
	"context"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type k8sJobEventHandler struct {
}

func (k *k8sJobEventHandler) Create(ctx context.Context, event event.CreateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (k *k8sJobEventHandler) Update(ctx context.Context, event event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (k *k8sJobEventHandler) Delete(ctx context.Context, event event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (k *k8sJobEventHandler) Generic(ctx context.Context, event event.GenericEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

type volcanoJobEventHandler struct {
}

func (v *volcanoJobEventHandler) Create(ctx context.Context, createEvent event.CreateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (v *volcanoJobEventHandler) Update(ctx context.Context, updateEvent event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (v *volcanoJobEventHandler) Delete(ctx context.Context, deleteEvent event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (v *volcanoJobEventHandler) Generic(ctx context.Context, genericEvent event.GenericEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

type serviceEventHandler struct {
}

func (s *serviceEventHandler) Create(ctx context.Context, createEvent event.CreateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceEventHandler) Update(ctx context.Context, updateEvent event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceEventHandler) Delete(ctx context.Context, deleteEvent event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceEventHandler) Generic(ctx context.Context, genericEvent event.GenericEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

type configmapEventHandler struct {
}

func (c *configmapEventHandler) Create(ctx context.Context, createEvent event.CreateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (c *configmapEventHandler) Update(ctx context.Context, updateEvent event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (c *configmapEventHandler) Delete(ctx context.Context, deleteEvent event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (c *configmapEventHandler) Generic(ctx context.Context, genericEvent event.GenericEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

type secretEventHandler struct {
}

func (s *secretEventHandler) Create(ctx context.Context, createEvent event.CreateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *secretEventHandler) Update(ctx context.Context, updateEvent event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *secretEventHandler) Delete(ctx context.Context, deleteEvent event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (s *secretEventHandler) Generic(ctx context.Context, genericEvent event.GenericEvent, queue workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}
