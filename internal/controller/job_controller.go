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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client

	// Clientset is a connection to the core kubernetes API
	//KubeClient *kubernetes.Clientset
	//VcClient   *vcclient.Clientset
	Cache *jobCache

	Scheme *runtime.Scheme
}

func NewJobReconciler(client client.Client, scheme *runtime.Scheme) (*JobReconciler, error) {

	r := &JobReconciler{
		Client: client,
		Scheme: scheme,
	}

	r.Cache = newJobCache()

	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	return nil, fmt.Errorf("build kube config err: %s", err.Error())
	//}
	//
	//r.KubeClient, err = kubernetes.NewForConfig(config)
	//if err != nil {
	//	return nil, fmt.Errorf("failed init kubeClient, with err: %v", err)
	//}
	//
	//r.VcClient, err = vcclient.NewForConfig(config)
	//if err != nil {
	//	return nil, fmt.Errorf("failed init vcClient, with err: %v", err)
	//}

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
		if errors.IsNotFound(err) {
			klog.Infof("Job resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}

		klog.Errorf("reconcile get job err: %s", err.Error())
		return ctrl.Result{}, fmt.Errorf("reconcile get job err: %s", err.Error())
	}

	// sync cache
	if err := r.Cache.syncGraphFromJob(job); err != nil {
		klog.Errorf("add job to tree cache err: %s", err.Error())
		return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
	}

	// if job was deleted, sync logic
	deletedFlag, err := r.Cache.isJobDeleted(job.Name)
	if err != nil {
		klog.Errorf(err.Error())
		return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
	}
	if deletedFlag {
		switch job.Status.State.Phase {
		case appsv1alpha1.Terminating:
			job.Status.State.Phase = appsv1alpha1.Terminated
			job.Status.State.Message = "job deleted"
			if err := r.updateJobStatus(context.Background(), job); err != nil {
				klog.Errorf(err.Error())
				return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
			}

			return ctrl.Result{}, nil
		case appsv1alpha1.Terminated:
			if err := r.deleteJob(context.Background(), job); err != nil {
				klog.Errorf(err.Error())
				return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
			}

			return ctrl.Result{}, nil
		default:
			job.Status.State.Phase = appsv1alpha1.Terminating
			job.Status.State.Message = "job deleting"
			if err := r.updateJobStatus(context.Background(), job); err != nil {
				klog.Errorf(err.Error())
				return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
			}

			return ctrl.Result{}, nil
		}

	}

	// if job was new created, update status
	if job.Status.State.Phase == appsv1alpha1.Unknown {
		job.Status.State.Phase = appsv1alpha1.Scheduled
		job.Status.State.Message = "job scheduled"
		job.Status.ItemStatus = map[string]appsv1alpha1.ItemStatus{}

		if err := r.updateJobStatus(context.Background(), job); err != nil {
			klog.Errorf(err.Error())
			return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
		}

		return ctrl.Result{}, nil
	}

	// job items' status
	changed, err := r.Cache.syncJobItemStatus(job)
	if err != nil {
		klog.Errorf(err.Error())
		return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
	}
	if changed {
		if err := r.updateJobStatus(context.Background(), job); err != nil {
			klog.Errorf(err.Error())
			return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
		}
	}

	// job finished
	finished, failed, err := r.Cache.isJobFinished(job.Name)
	if err != nil {
		klog.Errorf(err.Error())
		return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
	}
	if finished {
		switch job.Status.State.Phase {
		case appsv1alpha1.Terminated, appsv1alpha1.Terminating, appsv1alpha1.Failed, appsv1alpha1.Completed:
		case appsv1alpha1.Completing:
			job.Status.State.Phase = appsv1alpha1.Completed
			job.Status.State.Message = "job completed"

			if err := r.updateJobStatus(context.Background(), job); err != nil {
				klog.Errorf(err.Error())
				return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
			}

			return ctrl.Result{}, nil
		default:
			if failed {
				job.Status.State.Phase = appsv1alpha1.Failed
				job.Status.State.Message = "job failed"
			} else {
				job.Status.State.Phase = appsv1alpha1.Completing
				job.Status.State.Message = "job completing"
			}

			if err := r.updateJobStatus(context.Background(), job); err != nil {
				klog.Errorf(err.Error())
				return ctrl.Result{}, fmt.Errorf("reconcile job err: %s", err.Error())
			}

			return ctrl.Result{}, nil
		}
	}

	// if job was a scheduled one, schedule next items
	if job.Status.State.Phase == appsv1alpha1.Scheduled {
		if err := r.createJobItem(context.Background(), job); err != nil {
			klog.Errorf(err.Error())
			return ctrl.Result{}, fmt.Errorf("reconcile get job err: %s", err.Error())
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	filter := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if obj.GetObjectKind().GroupVersionKind().GroupVersion().Group == appsv1alpha1.GroupVersion.Group {
			return true
		}

		annotations := obj.GetAnnotations()
		_, ok := annotations[appsv1alpha1.CreateByJob]
		return ok
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Job{}).
		WithEventFilter(filter).
		WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		Watches(&v1.Job{}, handler.EnqueueRequestsFromMapFunc(r.Cache.kubeJobHandler)).
		Watches(&v1alpha1.Job{}, handler.EnqueueRequestsFromMapFunc(r.Cache.vcJobHandler)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(r.Cache.serviceHandler)).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.Cache.configmapHandler)).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.Cache.secretHandler)).
		Watches(&corev1.PersistentVolumeClaim{}, handler.EnqueueRequestsFromMapFunc(r.Cache.pvcHandler)).
		Watches(&corev1.PersistentVolume{}, handler.EnqueueRequestsFromMapFunc(r.Cache.pvHandler)).
		Complete(r)
}
