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

//+genclient
//+k8s:deepcopy-gen:package
//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=all,path=jobs,shortName=sfjob;sj
//+kubebuilder:subresource:status

// Job is the Schema for the jobs API
type Job struct {
	metav1.TypeMeta `json:",inline"`

	//+optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the songf job
	// +optional
	Spec JobSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Current status of the songf Job
	// +optional
	Status JobStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=spec"`
}

// JobSpec defines the desired state of Job
type JobSpec struct {

	// Items defines the specific step flow of the task, and based on this field,
	// a directed acyclic graph can be constructed to describe the task flow
	// +optional
	Items []Item `json:"items,omitempty" protobuf:"bytes,1,opt,name=items"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, components in all items will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,2,opt,name=ttlSecondsAfterFinished"`
}

// JobStatus defines the observed state of Job
type JobStatus struct {
	// Current state of Job.
	// +optional
	State JobState `json:"state,omitempty" protobuf:"bytes,1,opt,name=state"`

	// Current state of each open Item, including jobs and modules.
	// +optional
	ItemStatus map[string]ItemStatus `json:"itemStatus,omitempty" protobuf:"bytes,2,opt,name=itemStatus"`
}

// Item defines the specific execution process of Job
type Item struct {

	// The name of Item, must be Unique in all Items.
	// Can not set null.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Default to false.
	// If set true, this Item and its child Items won't participate in the scheduling of taskflow
	// +optional
	Truncated bool `json:"truncated,omitempty" protobuf:"varint,2,opt,name=truncated"`

	// RunAfter defines the timing of this Item can be scheduled.
	// When Items with name set in this field Success, this Item will start to run.
	// If set null, this Item will be the first one. Only one Item can set this field null.
	// +optional
	RunAfter []string `json:"runAfter,omitempty" protobuf:"bytes,3,opt,name=runAfter"`

	// ItemJobs defines the jobs scheduled in this Item, including volcano job and kube job.
	// +optional
	ItemJobs ItemJobResource `json:"itemJobs,omitempty" protobuf:"bytes,4,opt,name=ItemJobs"`

	// ItemModules defines the modules in this Item, including service, configmap and secret.
	// +optional
	ItemModules ItemModuleResource `json:"itemModules,omitempty" protobuf:"bytes,5,opt,name=itemModules"`
}

// ItemJobResource defines the jobs to create in Item
type ItemJobResource struct {

	// Jobs to create, the names of job must be unique.
	// +optional
	Jobs []ItemJobTemplate `json:"jobs,omitempty" protobuf:"bytes,1,opt,name=jobs"`
}

// ItemModuleResource defines the modules to create in Item
type ItemModuleResource struct {

	// ttlSecondsAfterFinished limits the lifetime of item modules that Item finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Item finishes, modules in all items will be
	// automatically deleted. If this field is unset,
	// the modules won't be automatically deleted. If this field is set to zero,
	// the modules becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// services to create, names can not be repeated
	// +optional
	Services []ServiceTemplate `json:"services,omitempty" protobuf:"bytes,2,opt,name=services"`

	// configmaps to create, names can not be repeated
	// +optional
	ConfigMaps []ConfigMapTemplate `json:"configMaps,omitempty" protobuf:"bytes,3,opt,name=configMaps"`

	// secrets to create, names can not be repeated
	// +optional
	Secrets []SecretTemplate `json:"secrets,omitempty" protobuf:"bytes,4,opt,name=secrets"`

	// pvcs to create, names can not be repeated
	// +optional
	Pvcs []PvcTemplate `json:"pvcs,omitempty" protobuf:"bytes,5,opt,name=pvcs"`

	// pvs to create, names can not be repeated
	// +optional
	Pvs []PvTemplate `json:"pvs,omitempty" protobuf:"bytes,6,opt,name=pvs"`
}

// ItemJobTemplate defines the jobs to create in Item, detailed information. KubeJobSpec and VolcanoJobSpec only one exist.
type ItemJobTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// If set, the container of the job will be replaced by the field.
	// For example, set this filed "a->b", it means the job's container will be replaced by Job with
	// name "b" that in Item with name "a".
	// +optional
	ContainerExtend *string `json:"containerExtend,omitempty" protobuf:"bytes,1,opt,name=containerExtend"`

	// If set, the pod of job will run on the node depends on the field.
	// For example, set this filed "a->b", it means the job's pods will run on the node that KubeJob with
	// name "b" that in Item with name "a" last finished.
	// For example, set this filed "a->b->c", it means the job's pods will run on the node that VolcanoJob with
	// name "b" that in Item with name "a" and has Task named "c" last finished.
	// If set, the follow job will only be allowed to run one task and one pod.
	// +optional
	NodeNameExtend *string `json:"nodeNameExtend,omitempty" protobuf:"bytes,2,opt,name=nodeNameExtend"`

	// Save container. If set true, job's container will be saved.
	// If other Item extend this job's container, will be auto set true.
	// +optional
	ContainerSave bool `json:"containerSave,omitempty" protobuf:"varint,3,opt,name=containerSave"`

	// Specification of the desired behavior of the job.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	KubeJobSpec *batchv1.JobSpec `json:"kubeJobSpec,omitempty" protobuf:"bytes,4,opt,name=kubeJobSpec"`

	// Specification of the desired behavior of the volcano job, including the minAvailable
	// +optional
	VolcanoJobSpec *v1alpha1.JobSpec `json:"VolcanoJobSpec,omitempty" protobuf:"bytes,5,opt,name=VolcanoJobSpec"`
}

// ServiceTemplate defines the service to create in Item, detailed information.
type ServiceTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, this service will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// Specification of the desired behavior of the volcano job, including the minAvailable
	// +optional
	Spec corev1.ServiceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ConfigMapTemplate defines the configmap to create in Item, detailed information.
type ConfigMapTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, this configmap will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// +optional
	ConfigMap corev1.ConfigMap `json:"configMap,omitempty" protobuf:"bytes,2,opt,name=configMap"`
}

// SecretTemplate defines the secret to create in Item, detailed information.
type SecretTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, this secret will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// +optional
	Secret corev1.Secret `json:"secret,omitempty" protobuf:"bytes,2,opt,name=secret"`
}

// PvcTemplate defines the pvc to create in Item, detailed information.
type PvcTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, this secret will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// +optional
	Pvc corev1.PersistentVolumeClaimSpec `json:"pvc,omitempty" protobuf:"bytes,2,opt,name=pvc"`
}

// PvTemplate defines the pv to create in Item, detailed information.
type PvTemplate struct {

	// +optional
	TemplateBaseInfo `json:",inline"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Completed or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, this secret will be
	// automatically deleted. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,1,opt,name=ttlSecondsAfterFinished"`

	// +optional
	Pv corev1.PersistentVolumeSpec `json:"pvc,omitempty" protobuf:"bytes,2,opt,name=pv"`
}

type TemplateBaseInfo struct {
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,2,rep,name=labels"`

	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,3,rep,name=annotations"`
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

// JobState defines the state of the job.
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

// ItemPhase defines the phase of the Item.
type ItemPhase string

const (
	ItemPending    ItemPhase = "Pending"
	ItemScheduling ItemPhase = "Scheduling"
	ItemScheduled  ItemPhase = "Scheduled"
	ItemCompleted  ItemPhase = "Completed"
	ItemFailed     ItemPhase = "Failed"
)

// ItemStatus defines the state of the item.
type ItemStatus struct {
	// The name of Item
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// The phase of Item.
	// +optional
	Phase ItemPhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase"`

	// The num of Job which is running.
	// +optional
	RunningJobNum *int32 `json:"runningJobNum,omitempty" protobuf:"bytes,3,opt,name=runningJobNum"`

	// The num of Job which is completed.
	// +optional
	CompletedJobNum *int32 `json:"completedJobNum,omitempty" protobuf:"bytes,4,opt,name=completedJobNum"`

	// The num of Job which is failed.
	// +optional
	FailedJobNum *int32 `json:"failedJobNum,omitempty" protobuf:"bytes,5,opt,name=failedJobNum"`

	// The status of volcano job, key is job name.
	// +optional
	JobStatus map[string]v1alpha1.JobState `json:"jobStatus,omitempty" protobuf:"bytes,6,opt,name=jobStatus"`

	// The status of service, key is service name.
	// +optional
	ServiceStatus map[string]RegularModuleStatus `json:"serviceStatus,omitempty" protobuf:"bytes,7,opt,name=serviceStatus"`

	// The status of configmap, key is configmap name.
	// +optional
	ConfigMapStatus map[string]RegularModuleStatus `json:"configMapStatus,omitempty" protobuf:"bytes,8,opt,name=configMapStatus"`

	// The status of secret, key is secret name.
	// +optional
	SecretStatus map[string]RegularModuleStatus `json:"secretStatus,omitempty" protobuf:"bytes,9,opt,name=secretStatus"`

	// +optional
	PvcStatus map[string]RegularModuleStatus `json:"pvcStatus,omitempty" protobuf:"bytes,9,opt,name=pvcStatus"`

	// +optional
	PvStatus map[string]RegularModuleStatus `json:"pvStatus,omitempty" protobuf:"bytes,9,opt,name=pvStatus"`
}

// RegularModulePhase defines the phase of regular module.
type RegularModulePhase string

const (
	RegularModuleUnknown  RegularModulePhase = "Unknown"
	RegularModuleCreating RegularModulePhase = "Creating"
	RegularModuleCreated  RegularModulePhase = "Created"
	RegularModuleFailed   RegularModulePhase = "Failed"
)

// RegularModuleStatus describe the status of module which don't need special description.
type RegularModuleStatus struct {
	// The phase of RegularModule.
	// +optional
	Phase RegularModulePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`

	// Last time the condition transit from one phase to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,2,opt,name=lastTransitionTime"`
}

type JobTemplateSpec struct {

	//+optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the songf job
	// +optional
	Spec JobSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}
