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
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	appsv1alpha1 "songf.sh/songf/pkg/api/apps.songf.sh/v1alpha1"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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

	if err := Cache.syncJobTree(job); err != nil {
		klog.Errorf("add job to tree cache err: %s", err.Error())
		return ctrl.Result{
			Requeue: true,
		}, fmt.Errorf("reconcile get job err: %s", err.Error())
	}

	return ctrl.Result{}, nil
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
