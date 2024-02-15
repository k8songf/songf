package controller

import (
	"context"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func (r *JobReconciler) createJobItem(ctx context.Context, job *appsv1alpha1.Job) error {

	schedulingItems, ok := r.Cache.getNextScheduleJobItem(job.Name)
	if !ok {
		klog.Errorf("create job item err: not find %s first item", job.Name)
	}

	for _, item := range schedulingItems {
		if err := r.createJobItemImpl(ctx, job, item); err != nil {
			return fmt.Errorf("create job item err: %s", err.Error())
		}
	}

	return nil
}

func (r *JobReconciler) createJobItemImpl(ctx context.Context, job *appsv1alpha1.Job, item *appsv1alpha1.Item) error {

	job.Status.ItemStatus[item.Name] = appsv1alpha1.ItemStatus{
		Name:  item.Name,
		Phase: appsv1alpha1.ItemScheduling,
	}

	baseAnnotations := job.Annotations
	baseAnnotations[appsv1alpha1.CreateByJob] = job.Name
	baseAnnotations[appsv1alpha1.CreateByJobItem] = item.Name

	baseLabels := job.Labels
	baseLabels[appsv1alpha1.CreateByJob] = job.Name
	baseLabels[appsv1alpha1.CreateByJobItem] = item.Name

	expendAnnotationFn := func(extend map[string]string) map[string]string {
		res := map[string]string{}

		for k, v := range baseAnnotations {
			res[k] = v
		}

		for k, v := range extend {
			res[k] = v
		}

		return res
	}

	expendLabelFn := func(extend map[string]string) map[string]string {
		res := map[string]string{}

		for k, v := range baseLabels {
			res[k] = v
		}

		for k, v := range extend {
			res[k] = v
		}

		return res
	}

	var createdObj []client.Object

	defer func() {
		for _, obj := range createdObj {
			if err := r.Delete(ctx, obj); err != nil {
				klog.Errorf(err.Error())
			}
		}
	}()

	for _, itemJob := range item.ItemJobs.Jobs {
		// todo container extend and node name apply

		if itemJob.KubeJobSpec == nil && itemJob.VolcanoJobSpec == nil {
			return fmt.Errorf("%s k8s itemJob and volcano itemJob can not be total nil", itemJob.Name)
		}

		if itemJob.KubeJobSpec != nil && itemJob.VolcanoJobSpec != nil {
			return fmt.Errorf("%s k8s itemJob and volcano itemJob can not be total exists", itemJob.Name)
		}

		jobName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, itemJob.Name)
		jobObjectMeta := metav1.ObjectMeta{
			Name:        jobName,
			Annotations: expendAnnotationFn(itemJob.Annotations),
			Labels:      expendLabelFn(itemJob.Labels),
		}

		var job2Create client.Object
		if itemJob.KubeJobSpec != nil {

			job2Create = &v1.Job{
				ObjectMeta: jobObjectMeta,
				Spec:       *itemJob.KubeJobSpec,
			}

		} else if itemJob.VolcanoJobSpec != nil {

			job2Create = &v1alpha1.Job{
				ObjectMeta: jobObjectMeta,
				Spec:       *itemJob.VolcanoJobSpec,
			}

		}

		if err := controllerutil.SetControllerReference(job, job2Create, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, job2Create); err != nil {
			return err
		}

		createdObj = append(createdObj, job2Create)

	}

	for _, service := range item.ItemModules.Services {

		serviceName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, service.Name)

		serviceObjectMeta := metav1.ObjectMeta{
			Name:        serviceName,
			Annotations: expendAnnotationFn(service.Annotations),
			Labels:      expendLabelFn(service.Labels),
		}

		service2Create := &corev1.Service{
			ObjectMeta: serviceObjectMeta,
			Spec:       service.Spec,
		}

		if err := controllerutil.SetControllerReference(job, service2Create, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, service2Create); err != nil {
			return err
		}

		createdObj = append(createdObj, service2Create)

	}

	for _, cm := range item.ItemModules.ConfigMaps {

		cmName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, cm.Name)

		cmImpl := cm.ConfigMap.DeepCopy()

		cmImpl.Labels = expendLabelFn(cm.Labels)
		cmImpl.Annotations = expendAnnotationFn(cm.Annotations)
		cmImpl.Name = cmName

		if err := controllerutil.SetControllerReference(job, cmImpl, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, cmImpl); err != nil {
			return err
		}

		createdObj = append(createdObj, cmImpl)

	}

	for _, secret := range item.ItemModules.Secrets {

		secretName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, secret.Name)

		secretImpl := secret.Secret.DeepCopy()

		secretImpl.Labels = expendLabelFn(secret.Labels)
		secretImpl.Annotations = expendAnnotationFn(secret.Annotations)
		secretImpl.Name = secretName

		if err := controllerutil.SetControllerReference(job, secretImpl, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, secretImpl); err != nil {
			return err
		}

		createdObj = append(createdObj, secretImpl)

	}

	for _, pv := range item.ItemModules.Pvs {

		pvName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, pv.Name)

		pvObjectMeta := metav1.ObjectMeta{
			Name:        pvName,
			Annotations: expendAnnotationFn(pv.Annotations),
			Labels:      expendLabelFn(pv.Labels),
		}

		pv2Create := &corev1.PersistentVolume{
			ObjectMeta: pvObjectMeta,
			Spec:       pv.Pv,
		}

		if err := controllerutil.SetControllerReference(job, pv2Create, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, pv2Create); err != nil {
			return err
		}

		createdObj = append(createdObj, pv2Create)

	}

	for _, pvc := range item.ItemModules.Pvcs {

		pvcName := appsv1alpha1.CalJobItemSubName(job.Name, item.Name, pvc.Name)

		pvcObjectMeta := metav1.ObjectMeta{
			Name:        pvcName,
			Annotations: expendAnnotationFn(pvc.Annotations),
			Labels:      expendLabelFn(pvc.Labels),
		}

		pvc2Create := &corev1.PersistentVolumeClaim{
			ObjectMeta: pvcObjectMeta,
			Spec:       pvc.Pvc,
		}

		if err := controllerutil.SetControllerReference(job, pvc2Create, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, pvc2Create); err != nil {
			return err
		}

		createdObj = append(createdObj, pvc2Create)

	}

	return nil

}

func (r *JobReconciler) updateJobStatus(ctx context.Context, job *appsv1alpha1.Job) error {
	if err := r.Client.Status().Update(ctx, job); err != nil {
		return fmt.Errorf("update job status while delete err: %s", err.Error())
	}

	return nil
}

func (r *JobReconciler) deleteJob(ctx context.Context, job *appsv1alpha1.Job) error {
	// todo check sub items finished

	if err := r.Client.Delete(context.Background(), job); err != nil {
		return fmt.Errorf("delete job err: %s", err.Error())
	}

	return nil
}
