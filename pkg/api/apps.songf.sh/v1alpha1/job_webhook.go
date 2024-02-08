package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-songf-sh-v1alpha1-job,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.songf.sh,resources=jobs,verbs=create;update,versions=v1alpha1,name=mjob.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Job{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Job) Default() {
	klog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-songf-sh-v1alpha1-job,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.songf.sh,resources=jobs,verbs=create;update,versions=v1alpha1,name=vjob.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Job{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateCreate() (admission.Warnings, error) {
	klog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	klog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateDelete() (admission.Warnings, error) {
	klog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
