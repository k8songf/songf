package controller

import (
	"context"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func (c *jobCache) kubeJobHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := getJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	job, ok := object.(*v1.Job)
	if !ok {
		klog.Errorf("receive object %v/%v which is not kubeJob", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		tree = newJobItemTree()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		jobState, ok := status.JobStatus[job.Name]
		if !ok {
			jobState = v1alpha1.JobState{
				Phase:              v1alpha1.Pending,
				LastTransitionTime: job.CreationTimestamp,
			}
		}

		if job.DeletionTimestamp != nil && !job.DeletionTimestamp.IsZero() {
			jobState = v1alpha1.JobState{
				Phase:              v1alpha1.Terminated,
				LastTransitionTime: *job.DeletionTimestamp,
			}
			status.JobStatus[job.Name] = jobState
			return
		}

		conditionsLength := len(job.Status.Conditions)
		if conditionsLength > 0 {
			condition := job.Status.Conditions[conditionsLength-1]

			jobState.LastTransitionTime = condition.LastTransitionTime
			jobState.Message = condition.Message
			jobState.Reason = condition.Reason

			switch condition.Type {
			case v1.JobSuspended:
				jobState.Phase = v1alpha1.Aborted

			case v1.JobComplete:
				jobState.Phase = v1alpha1.Completed

			case v1.JobFailureTarget, v1.JobFailed:
				jobState.Phase = v1alpha1.Failed

			default:
				klog.Errorf("can not recognize kube job %s/%s condition type: %s", job.Namespace, job.Name, condition.Type)
			}
		}

		status.JobStatus[job.Name] = jobState
		return

	}

	if err := tree.syncFromObject(job, fn); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	c.jobItemTreeCache[jobName] = tree

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

	jobName, _ := getJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	job, ok := object.(*v1alpha1.Job)
	if !ok {
		klog.Errorf("receive object %v/%v which is not vcJob", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		tree = newJobItemTree()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		status.JobStatus[job.Name] = job.Status.State
	}

	if err := tree.syncFromObject(job, fn); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	c.jobItemTreeCache[jobName] = tree

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

	jobName, _ := getJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	service, ok := object.(*corev1.Service)
	if !ok {
		klog.Errorf("receive object %v/%v which is not service", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		tree = newJobItemTree()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		serviceStatus, ok := status.ServiceStatus[service.Name]
		if !ok {
			serviceStatus = appsv1alpha1.RegularModuleStatus{
				Phase: appsv1alpha1.RegularModuleUnknown,
			}
		}

		if service.DeletionTimestamp == nil || service.DeletionTimestamp.IsZero() {

			serviceStatus.Phase = appsv1alpha1.RegularModuleCreated
			serviceStatus.LastTransitionTime = service.CreationTimestamp
		} else {
			serviceStatus.Phase = appsv1alpha1.RegularModuleFailed
			serviceStatus.LastTransitionTime = *service.DeletionTimestamp
		}
		status.ServiceStatus[service.Name] = serviceStatus
	}

	if err := tree.syncFromObject(service, fn); err != nil {
		klog.Errorf("%s/%s sync tree from cache err: %s", service.Namespace, service.Name, err.Error())
		return nil
	}

	c.jobItemTreeCache[jobName] = tree

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

	jobName, _ := getJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	configmap, ok := object.(*corev1.ConfigMap)
	if !ok {
		klog.Errorf("receive object %v/%v which is not secret", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		tree = newJobItemTree()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		cmStatus, ok := status.ConfigMapStatus[configmap.Name]
		if !ok {
			cmStatus = appsv1alpha1.RegularModuleStatus{
				Phase: appsv1alpha1.RegularModuleUnknown,
			}
		}

		if configmap.DeletionTimestamp == nil || configmap.DeletionTimestamp.IsZero() {

			cmStatus.Phase = appsv1alpha1.RegularModuleCreated
			cmStatus.LastTransitionTime = configmap.CreationTimestamp
		} else {
			cmStatus.Phase = appsv1alpha1.RegularModuleFailed
			cmStatus.LastTransitionTime = *configmap.DeletionTimestamp
		}
		status.ServiceStatus[configmap.Name] = cmStatus
	}

	if err := tree.syncFromObject(configmap, fn); err != nil {
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

	jobName, _ := getJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	secret, ok := object.(*corev1.Secret)
	if !ok {
		klog.Errorf("receive object %v/%v which is not configmap", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	tree, ok := c.jobItemTreeCache[jobName]
	if !ok {
		tree = newJobItemTree()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		secretStatus, ok := status.SecretStatus[secret.Name]
		if !ok {
			secretStatus = appsv1alpha1.RegularModuleStatus{
				Phase: appsv1alpha1.RegularModuleUnknown,
			}
		}

		if secret.DeletionTimestamp == nil || secret.DeletionTimestamp.IsZero() {

			secretStatus.Phase = appsv1alpha1.RegularModuleCreated
			secretStatus.LastTransitionTime = secret.CreationTimestamp
		} else {
			secretStatus.Phase = appsv1alpha1.RegularModuleFailed
			secretStatus.LastTransitionTime = *secret.DeletionTimestamp
		}
		status.SecretStatus[secret.Name] = secretStatus
	}

	if err := tree.syncFromObject(secret, fn); err != nil {
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
