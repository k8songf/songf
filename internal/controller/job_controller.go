/*
Copyright 2023 firewood.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	vcclient "volcano.sh/apis/pkg/client/clientset/versioned"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client

	// Clientset is a connection to the core kubernetes API
	KubeClient *kubernetes.Clientset
	VcClient   *vcclient.Clientset

	Scheme *runtime.Scheme
}

func NewJobReconciler(client client.Client, scheme *runtime.Scheme) (*JobReconciler, error) {

	r := &JobReconciler{
		Client: client,
		Scheme: scheme,
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("build kube config err: %s", err.Error())
	}

	r.KubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed init kubeClient, with err: %v", err)
	}

	r.VcClient, err = vcclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed init vcClient, with err: %v", err)
	}

	return r, nil
}

//+kubebuilder:rbac:groups=apps.songf.sh,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.songf.sh,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.songf.sh,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Job object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.2/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	job := &appsv1alpha1.Job{}

	if err := r.Client.Get(ctx, req.NamespacedName, job); err != nil {
		klog.Errorf("reconcile get job err: %s", err.Error())
		return ctrl.Result{
			Requeue: true,
		}, fmt.Errorf("reconcile get job err: %s", err.Error())
	}

	schedulingItems, err := Cache.syncJobTree(job)
	if err != nil {
		klog.Errorf("add job to tree cache err: %s", err.Error())
		return ctrl.Result{
			Requeue: true,
		}, fmt.Errorf("reconcile get job err: %s", err.Error())
	}

	switch job.Status.State.Phase {
	case appsv1alpha1.Unknown:
		job.Status.State.Phase = appsv1alpha1.Scheduled
		job.Status.State.Message = "job scheduled"

		job.Status.ItemStatus = map[string]appsv1alpha1.ItemStatus{}
		for _, item := range job.Spec.Items {

			if len(item.RunAfter) == 0 {
				job.Status.ItemStatus[item.Name] = appsv1alpha1.ItemStatus{
					Name:  item.Name,
					Phase: appsv1alpha1.ItemScheduling,
				}

				if err = r.createJobItem(context.Background(), job, &item); err != nil {
					klog.Errorf("create job item err: %s", err.Error())
					return ctrl.Result{
						Requeue: true,
					}, fmt.Errorf("reconcile get job err: %s", err.Error())
				}

			} else {
				job.Status.ItemStatus[item.Name] = appsv1alpha1.ItemStatus{
					Name:  item.Name,
					Phase: appsv1alpha1.ItemPending,
				}
			}

		}

		if err = r.Client.Status().Update(context.Background(), job); err != nil {
			klog.Errorf("add job to tree cache err: %s", err.Error())
			return ctrl.Result{
				Requeue: true,
			}, fmt.Errorf("reconcile get job err: %s", err.Error())
		}
	default:
		for _, node := range schedulingItems {
			job.Status.ItemStatus[node.item.Name] = appsv1alpha1.ItemStatus{
				Name:  node.item.Name,
				Phase: appsv1alpha1.ItemScheduling,
			}

			if err = r.createJobItem(context.Background(), job, node.item); err != nil {
				klog.Errorf("create job item err: %s", err.Error())
				return ctrl.Result{
					Requeue: true,
				}, fmt.Errorf("reconcile get job err: %s", err.Error())
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *JobReconciler) createJobItem(ctx context.Context, job *appsv1alpha1.Job, item *appsv1alpha1.Item) error {

	trueFlag := true
	ownerReference := []metav1.OwnerReference{
		{
			APIVersion:         job.APIVersion,
			Kind:               job.Kind,
			Name:               job.Name,
			UID:                job.UID,
			Controller:         &trueFlag,
			BlockOwnerDeletion: &trueFlag,
		},
	}

	jobAnnotations := job.Annotations
	jobLabels := job.Labels

	expendAnnotationFn := func(extend map[string]string) map[string]string {
		res := map[string]string{}

		for k, v := range jobAnnotations {
			res[k] = v
		}

		for k, v := range extend {
			res[k] = v
		}

		return res
	}

	expendLabelFn := func(extend map[string]string) map[string]string {
		res := map[string]string{}

		for k, v := range jobLabels {
			res[k] = v
		}

		for k, v := range extend {
			res[k] = v
		}

		return res
	}

	var vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated []string
	releaseFn := func(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated []string) {
		for _, name := range vcJobCreated {
			if err := r.KubeClient.BatchV1().Jobs(metav1.NamespaceDefault).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf(err.Error())
			}
		}

		for _, name := range kubeJobCreated {
			if err := r.VcClient.BatchV1alpha1().Jobs(metav1.NamespaceDefault).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf(err.Error())
			}
		}

		for _, name := range serviceCreated {
			if err := r.KubeClient.CoreV1().Services(metav1.NamespaceDefault).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf(err.Error())
			}
		}

		for _, name := range cmCreated {
			if err := r.KubeClient.CoreV1().ConfigMaps(metav1.NamespaceDefault).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf(err.Error())
			}
		}

		for _, name := range secretCreated {
			if err := r.KubeClient.CoreV1().Secrets(metav1.NamespaceDefault).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf(err.Error())
			}
		}
	}

	for _, itemJob := range item.ItemJobs.Jobs {
		// todo container extend and node name apply

		if itemJob.K8sJobSpec == nil && itemJob.VolcanoJobSpec == nil {
			releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
			return fmt.Errorf("%s k8s itemJob and volcano itemJob can not be total nil", itemJob.Name)
		}

		if itemJob.K8sJobSpec != nil && itemJob.VolcanoJobSpec != nil {
			releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
			return fmt.Errorf("%s k8s itemJob and volcano itemJob can not be total exists", itemJob.Name)
		}

		jobName := fmt.Sprintf("%s-%s", item.Name, itemJob.Name)
		jobObjectMeta := metav1.ObjectMeta{
			Name:            jobName,
			OwnerReferences: ownerReference,
			Annotations:     expendAnnotationFn(itemJob.Annotations),
			Labels:          expendLabelFn(itemJob.Labels),
		}

		if itemJob.K8sJobSpec != nil {

			job2Create := &v1.Job{
				ObjectMeta: jobObjectMeta,
				Spec:       *itemJob.K8sJobSpec,
			}

			_, err := r.KubeClient.BatchV1().Jobs(metav1.NamespaceDefault).Create(ctx, job2Create, metav1.CreateOptions{})
			if err != nil {
				releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
				return err
			}

			kubeJobCreated = append(kubeJobCreated, jobName)

			continue
		}

		if itemJob.VolcanoJobSpec != nil {

			job2Create := &v1alpha1.Job{
				ObjectMeta: jobObjectMeta,
				Spec:       *itemJob.VolcanoJobSpec,
			}

			_, err := r.VcClient.BatchV1alpha1().Jobs(metav1.NamespaceDefault).Create(ctx, job2Create, metav1.CreateOptions{})
			if err != nil {
				releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
				return err
			}

			vcJobCreated = append(vcJobCreated, jobName)

			continue
		}

	}

	for _, service := range item.ItemModules.Services {

		serviceName := fmt.Sprintf("%s-%s", item.Name, service.Name)

		serviceObjectMeta := metav1.ObjectMeta{
			Name:            serviceName,
			OwnerReferences: ownerReference,
			Annotations:     expendAnnotationFn(service.Annotations),
			Labels:          expendLabelFn(service.Labels),
		}

		service2Create := &corev1.Service{
			ObjectMeta: serviceObjectMeta,
			Spec:       service.Spec,
		}

		_, err := r.KubeClient.CoreV1().Services(metav1.NamespaceDefault).Create(ctx, service2Create, metav1.CreateOptions{})
		if err != nil {
			releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
			return err
		}

		serviceCreated = append(serviceCreated, serviceName)

	}

	for _, cm := range item.ItemModules.ConfigMaps {

		cmName := fmt.Sprintf("%s-%s", item.Name, cm.ConfigMap.Name)

		cmImpl := cm.ConfigMap.DeepCopy()

		cmImpl.Labels = expendLabelFn(cm.ConfigMap.Labels)
		cmImpl.Annotations = expendAnnotationFn(cm.ConfigMap.Annotations)
		cmImpl.OwnerReferences = ownerReference
		cmImpl.Name = cmName

		_, err := r.KubeClient.CoreV1().ConfigMaps(metav1.NamespaceDefault).Create(ctx, cmImpl, metav1.CreateOptions{})
		if err != nil {
			releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
			return err
		}

		cmCreated = append(cmCreated, cmName)

	}

	for _, secret := range item.ItemModules.Secrets {

		secretName := fmt.Sprintf("%s-%s", item.Name, secret.Secret.Name)

		secretImpl := secret.Secret.DeepCopy()

		secretImpl.Labels = expendLabelFn(secret.Secret.Labels)
		secretImpl.Annotations = expendAnnotationFn(secret.Secret.Annotations)
		secretImpl.OwnerReferences = ownerReference
		secretImpl.Name = secretName

		_, err := r.KubeClient.CoreV1().Secrets(metav1.NamespaceDefault).Create(ctx, secretImpl, metav1.CreateOptions{})
		if err != nil {
			releaseFn(vcJobCreated, kubeJobCreated, serviceCreated, cmCreated, secretCreated)
			return err
		}

		secretCreated = append(secretCreated, secretName)

	}

	return nil

}

const CreateByJob = "songf.sh/job"

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	filter := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if obj.GetObjectKind().GroupVersionKind().GroupVersion().Group == appsv1alpha1.GroupVersion.Group {
			return true
		}

		annotations := obj.GetAnnotations()
		_, ok := annotations[CreateByJob]
		return ok
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Job{}).
		WithEventFilter(filter).
		WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		Watches(&v1.Job{}, &k8sJobEventHandler{}).
		Watches(&v1alpha1.Job{}, &volcanoJobEventHandler{}).
		Watches(&corev1.Service{}, &serviceEventHandler{}).
		Watches(&corev1.ConfigMap{}, &configmapEventHandler{}).
		Watches(&corev1.Secret{}, &secretEventHandler{}).
		Complete(r)
}
