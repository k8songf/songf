package v1alpha1

import (
	"fmt"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-apps-songf-sh-v1alpha1-job,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.songf.sh,resources=jobs,verbs=create;update,versions=v1alpha1,name=mjob.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Job{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Job) Default() {
	klog.Info("default", "name", r.Name)

	// container save set
	containerSaveMap := map[string]string{}

	for _, item := range r.Spec.Items {
		for _, jobTemplate := range item.ItemJobs.Jobs {
			if jobTemplate.ContainerExtend != nil && *jobTemplate.ContainerExtend != "" {
				names := JobExtendStr2Names(*jobTemplate.ContainerExtend)
				if len(names) < 2 {
					klog.Errorf("container extend err: length of %s at least 2", *jobTemplate.ContainerExtend)
					continue
				}

				containerSaveMap[names[0]] = names[1]
			}
		}
	}

	for i, item := range r.Spec.Items {
		jobName, ok := containerSaveMap[item.Name]
		if ok {
			for j, itemJob := range item.ItemJobs.Jobs {
				if itemJob.Name == jobName {
					r.Spec.Items[i].ItemJobs.Jobs[j].ContainerSave = true
					break
				}
			}
		}
	}

}

//+kubebuilder:webhook:path=/validate-apps-songf-sh-v1alpha1-job,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.songf.sh,resources=jobs,verbs=create;update;delete,versions=v1alpha1,name=vjob.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Job{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateCreate() (admission.Warnings, error) {
	klog.Info("validate create", "name", r.Name)

	var warnings admission.Warnings

	if len(r.Spec.Items) == 0 {
		errMsg := fmt.Sprintf("%s items can not be nil", r.Name)
		warnings = append(warnings, errMsg)
		return warnings, fmt.Errorf(errMsg)
	}

	if flag, msg := IsJobItemValid(r); !flag {
		warnings = append(warnings, msg)
		return warnings, fmt.Errorf(msg)
	}

	flag, err := IsJobHasCycle(r)
	if err != nil {
		warnings = append(warnings, err.Error())
		return warnings, err
	}
	if !flag {
		msg := "can not build cycle graph in job"
		warnings = append(warnings, msg)
		return warnings, fmt.Errorf(msg)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	klog.Info("validate update", "name", r.Name)

	var warnings admission.Warnings

	oldJob, ok := old.(*Job)
	if !ok {
		warnings = append(warnings, "update old is not Job")
	}

	if !apiequality.Semantic.DeepEqual(r.Spec, oldJob.Spec) {
		msg := "job updates may not change fields other than"
		warnings = append(warnings, msg)
		return warnings, fmt.Errorf(msg)
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Job) ValidateDelete() (admission.Warnings, error) {
	klog.Info("validate delete", "name", r.Name)

	return nil, nil
}
