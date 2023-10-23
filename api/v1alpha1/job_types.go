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

package v1alpha1

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+genclient
//+k8s:deepcopy-gen=package
//+kubebuilder:object:root=true
//+kubebuilder:resource:path=jobs,shortName=sfjob;sj
//+kubebuilder:subresource:status

// Job is the Schema for the jobs API
type Job struct {
	metav1.TypeMeta `json:",inline"`

	//+optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the songf job
	// +optional
	Spec JobSpec `json:"spec,omitempty"`

	// Current status of the songf Job
	// +optional
	Status JobStatus `json:"status,omitempty"`
}

// JobSpec defines the desired state of Job
type JobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Items []Item `json:"items,omitempty"`

	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,8,opt,name=ttlSecondsAfterFinished"`
}

// JobStatus defines the observed state of Job
type JobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	State JobState `json:"state,omitempty" protobuf:"bytes,1,opt,name=state"`

	ItemStatus map[string]ItemStatus
}

type Item struct {
	Name string

	Truncated bool

	RunAfter []string

	TotalDeleteWhileFail *bool

	ItemJobResource

	ItemModuleResource
}

type ItemJobResource struct {
	ContainerExtend *string

	TotalDeleteWhileFail *bool

	Jobs []ItemJobTemplate
}

type ItemModuleResource struct {
	CleanUp bool

	TotalDeleteWhileFail *bool

	Services []ServiceTemplate

	ConfigMap []ConfigMapTemplate

	Secret []corev1.Secret
}

type ItemJobTemplate struct {
	// Standard object's metadata of the jobs created from this template.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	ContainerExtend *string

	NodeNameExtend *string

	// Specification of the desired behavior of the job.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	K8sJobSpec *batchv1.JobSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Specification of the desired behavior of the volcano job, including the minAvailable
	// +optional
	VolcanoJobSpec *v1alpha1.JobSpec `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
}

type ServiceTemplate struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	CleanUp bool

	// Specification of the desired behavior of the volcano job, including the minAvailable
	// +optional
	Spec corev1.ServiceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type ConfigMapTemplate struct {
	CleanUp bool

	ConfigMap corev1.ConfigMap
}

type SecretTemplate struct {
	CleanUp bool

	Secret corev1.Secret
}

// JobPhase defines the phase of the job.
type JobPhase string

const (
	Unknown     JobPhase = "Unknown"
	Scheduled   JobPhase = "Scheduled"
	Completed   JobPhase = "Completed"
	Failed      JobPhase = "Failed"
	Completing  JobPhase = "Completing"
	Terminating JobPhase = "Terminating"
	Terminated  JobPhase = "Terminated"
)

type JobState struct {
	// The phase of Job.
	// +optional
	Phase JobPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`

	// Unique, one-word, CamelCase reason for the phase's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// Last time the condition transit from one phase to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
}

// JobPhase defines the phase of the job.
type ItemPhase string

const (
	ItemPending   RegularModulePhase = "Unknown"
	ItemScheduled RegularModulePhase = "Scheduled"
	ItemCompleted RegularModulePhase = "Completed"
	ItemFailed    RegularModulePhase = "Failed"
)

type ItemStatus struct {
	Name string

	Phase ItemPhase

	RunningJobNum *int32

	CompletedJobNum *int32

	FailedJobNum *int32

	K8sJobStatus map[string]v1alpha1.JobState

	VolcanoJobStatus map[string]v1alpha1.JobState

	ServiceStatus map[string]RegularModuleStatus

	ConfigMapStatus map[string]RegularModuleStatus

	SecretStatus map[string]RegularModuleStatus
}

// JobPhase defines the phase of the job.
type RegularModulePhase string

const (
	RegularModuleUnknown  RegularModulePhase = "Unknown"
	RegularModuleCreating RegularModulePhase = "Creating"
	RegularModuleCreated  RegularModulePhase = "Created"
	RegularModuleFailed   RegularModulePhase = "Failed"
)

type RegularModuleStatus struct {
	Phase              RegularModulePhase
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
}

type JobTemplateSpec struct {

	//+optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the songf job
	// +optional
	Spec JobSpec `json:"spec,omitempty"`
}

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}
