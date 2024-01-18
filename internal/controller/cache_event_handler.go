package controller

import (
	"context"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func (c *jobCache) kubeJobHandler(ctx context.Context, object client.Object) []reconcile.Request {

	job, ok := object.(*v1.Job)
	if !ok {
		klog.Errorf("receive object %v/%v which is not kubeJob", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, err := c.getTreeFromAnnotationsAndOwnerReferences(job.Annotations, job.OwnerReferences)
	if err != nil {
		klog.Errorf("%s/%s get tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	if err = tree.syncFromKubeJob(job); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tree.name,
				Namespace: tree.nameSpace,
			},
		},
	}

}

func (c *jobCache) vcJobHandler(ctx context.Context, object client.Object) []reconcile.Request {
	job, ok := object.(*v1alpha1.Job)
	if !ok {
		klog.Errorf("receive object %v/%v which is not vcJob", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, err := c.getTreeFromAnnotationsAndOwnerReferences(job.Annotations, job.OwnerReferences)
	if err != nil {
		klog.Errorf("%s/%s get tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	if err = tree.syncFromVcJob(job); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tree.name,
				Namespace: tree.nameSpace,
			},
		},
	}

}

func (c *jobCache) serviceHandler(ctx context.Context, object client.Object) []reconcile.Request {
	service, ok := object.(*corev1.Service)
	if !ok {
		klog.Errorf("receive object %v/%v which is not service", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, err := c.getTreeFromAnnotationsAndOwnerReferences(service.Annotations, service.OwnerReferences)
	if err != nil {
		klog.Errorf("%s/%s get tree from cache err: %s", service.Namespace, service.Name, err.Error())
		return nil
	}

	if err = tree.syncFromService(service); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", service.Namespace, service.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tree.name,
				Namespace: tree.nameSpace,
			},
		},
	}

}

func (c *jobCache) configmapHandler(ctx context.Context, object client.Object) []reconcile.Request {

	configmap, ok := object.(*corev1.ConfigMap)
	if !ok {
		klog.Errorf("receive object %v/%v which is not secret", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, err := c.getTreeFromAnnotationsAndOwnerReferences(configmap.Annotations, configmap.OwnerReferences)
	if err != nil {
		klog.Errorf("%s/%s get tree from cache err: %s", configmap.Namespace, configmap.Name, err.Error())
		return nil
	}

	if err = tree.syncFromConfigmap(configmap); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", configmap.Namespace, configmap.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tree.name,
				Namespace: tree.nameSpace,
			},
		},
	}

}

func (c *jobCache) secretHandler(ctx context.Context, object client.Object) []reconcile.Request {

	secret, ok := object.(*corev1.Secret)
	if !ok {
		klog.Errorf("receive object %v/%v which is not configmap", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, err := c.getTreeFromAnnotationsAndOwnerReferences(secret.Annotations, secret.OwnerReferences)
	if err != nil {
		klog.Errorf("%s/%s get tree from cache err: %s", secret.Namespace, secret.Name, err.Error())
		return nil
	}

	if err = tree.syncFromSecret(secret); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", secret.Namespace, secret.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tree.name,
				Namespace: tree.nameSpace,
			},
		},
	}

}

func (c *jobCache) getTreeFromAnnotationsAndOwnerReferences(annotations map[string]string, references []metav1.OwnerReference) (*jobItemTree, error) {
	jobName, ok := annotations[CreateByJob]
	if !ok {
		return nil, fmt.Errorf("not found job name from annotations")
	}

	var uuid types.UID
	for _, reference := range references {
		if reference.Name == jobName && reference.APIVersion == appsv1alpha1.GroupVersion.String() {
			uuid = reference.UID
		}
	}

	tree, ok := c.jobItemTreeCache[uuid]
	if !ok {
		return nil, fmt.Errorf("not found job tree from cache %s/%s", jobName, string(uuid))
	}

	return tree, nil
}
