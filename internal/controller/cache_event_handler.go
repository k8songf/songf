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
	"songf.sh/songf/pkg/api/utils"
	"songf.sh/songf/pkg/job_graph"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func (c *jobCache) kubeJobHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
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

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
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

	if err := graph.SyncFromObject(job, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	c.jobItemGraphCache[jobName] = graph

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) vcJobHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
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

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		status.JobStatus[job.Name] = job.Status.State
	}

	if err := graph.SyncFromObject(job, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", job.Namespace, job.Name, err.Error())
		return nil
	}

	c.jobItemGraphCache[jobName] = graph

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) serviceHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
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

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
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

	if err := graph.SyncFromObject(service, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", service.Namespace, service.Name, err.Error())
		return nil
	}

	c.jobItemGraphCache[jobName] = graph

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) configmapHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
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

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
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

	if err := graph.SyncFromObject(configmap, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", configmap.Namespace, configmap.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) secretHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	secret, ok := object.(*corev1.Secret)
	if !ok {
		klog.Errorf("receive object %v/%v which is not secret", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
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

	if err := graph.SyncFromObject(secret, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", secret.Namespace, secret.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) pvcHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	pvc, ok := object.(*corev1.PersistentVolumeClaim)
	if !ok {
		klog.Errorf("receive object %v/%v which is not pvc", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		pvcStatus, ok := status.PvcStatus[pvc.Name]
		if !ok {
			pvcStatus = appsv1alpha1.RegularModuleStatus{
				Phase: appsv1alpha1.RegularModuleUnknown,
			}
		}

		if pvc.DeletionTimestamp == nil || pvc.DeletionTimestamp.IsZero() {

			switch pvc.Status.Phase {
			case corev1.ClaimPending:
				pvcStatus.Phase = appsv1alpha1.RegularModuleCreating
				pvcStatus.LastTransitionTime = pvc.CreationTimestamp
			case corev1.ClaimBound:
				pvcStatus.Phase = appsv1alpha1.RegularModuleCreating
			case corev1.ClaimLost:
				pvcStatus.Phase = appsv1alpha1.RegularModuleFailed
			}
		} else {
			pvcStatus.Phase = appsv1alpha1.RegularModuleFailed
			pvcStatus.LastTransitionTime = *pvc.DeletionTimestamp
		}
		status.SecretStatus[pvc.Name] = pvcStatus
	}

	if err := graph.SyncFromObject(pvc, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", pvc.Namespace, pvc.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}

func (c *jobCache) pvHandler(ctx context.Context, object client.Object) []reconcile.Request {

	jobName, _ := utils.GetJobNameAndItemNameFromObject(object)
	if jobName == "" {
		klog.Errorf("receive object %v/%v which is not belong job", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	pv, ok := object.(*corev1.PersistentVolume)
	if !ok {
		klog.Errorf("receive object %v/%v which is not pv", object.GetObjectKind().GroupVersionKind().Kind, object.GetName())
		return nil
	}

	c.Lock()
	defer c.Unlock()

	graph, ok := c.jobItemGraphCache[jobName]
	if !ok {
		graph = job_graph.NewJobItemGraph()
	}

	fn := func(status *appsv1alpha1.ItemStatus) {
		pvStatus, ok := status.PvStatus[pv.Name]
		if !ok {
			pvStatus = appsv1alpha1.RegularModuleStatus{
				Phase: appsv1alpha1.RegularModuleUnknown,
			}
		}

		if pv.DeletionTimestamp == nil || pv.DeletionTimestamp.IsZero() {

			switch pv.Status.Phase {
			case corev1.VolumePending, corev1.VolumeAvailable:
				pvStatus.Phase = appsv1alpha1.RegularModuleCreating
				pvStatus.LastTransitionTime = pv.CreationTimestamp
			case corev1.VolumeBound:
				pvStatus.Phase = appsv1alpha1.RegularModuleCreating
			case corev1.VolumeFailed, corev1.VolumeReleased:
				pvStatus.Phase = appsv1alpha1.RegularModuleFailed
			}
		} else {
			pvStatus.Phase = appsv1alpha1.RegularModuleFailed
			pvStatus.LastTransitionTime = *pv.DeletionTimestamp
		}
		status.SecretStatus[pv.Name] = pvStatus
	}

	if err := graph.SyncFromObject(pv, fn); err != nil {
		klog.Errorf("%s/%s sync graph from cache err: %s", pv.Namespace, pv.Name, err.Error())
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      graph.Name,
				Namespace: graph.NameSpace,
			},
		},
	}

}
