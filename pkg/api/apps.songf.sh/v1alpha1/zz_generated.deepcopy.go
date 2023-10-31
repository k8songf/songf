//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/batch/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	batchv1alpha1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapTemplate) DeepCopyInto(out *ConfigMapTemplate) {
	*out = *in
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
	in.ConfigMap.DeepCopyInto(&out.ConfigMap)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapTemplate.
func (in *ConfigMapTemplate) DeepCopy() *ConfigMapTemplate {
	if in == nil {
		return nil
	}
	out := new(ConfigMapTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Item) DeepCopyInto(out *Item) {
	*out = *in
	if in.RunAfter != nil {
		in, out := &in.RunAfter, &out.RunAfter
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.ItemJobs.DeepCopyInto(&out.ItemJobs)
	in.ItemModules.DeepCopyInto(&out.ItemModules)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Item.
func (in *Item) DeepCopy() *Item {
	if in == nil {
		return nil
	}
	out := new(Item)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ItemJobResource) DeepCopyInto(out *ItemJobResource) {
	*out = *in
	if in.ContainerExtend != nil {
		in, out := &in.ContainerExtend, &out.ContainerExtend
		*out = new(string)
		**out = **in
	}
	if in.Jobs != nil {
		in, out := &in.Jobs, &out.Jobs
		*out = make([]ItemJobTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ItemJobResource.
func (in *ItemJobResource) DeepCopy() *ItemJobResource {
	if in == nil {
		return nil
	}
	out := new(ItemJobResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ItemJobTemplate) DeepCopyInto(out *ItemJobTemplate) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.ContainerExtend != nil {
		in, out := &in.ContainerExtend, &out.ContainerExtend
		*out = new(string)
		**out = **in
	}
	if in.NodeNameExtend != nil {
		in, out := &in.NodeNameExtend, &out.NodeNameExtend
		*out = new(string)
		**out = **in
	}
	if in.K8sJobSpec != nil {
		in, out := &in.K8sJobSpec, &out.K8sJobSpec
		*out = new(v1.JobSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.VolcanoJobSpec != nil {
		in, out := &in.VolcanoJobSpec, &out.VolcanoJobSpec
		*out = new(batchv1alpha1.JobSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ItemJobTemplate.
func (in *ItemJobTemplate) DeepCopy() *ItemJobTemplate {
	if in == nil {
		return nil
	}
	out := new(ItemJobTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ItemModuleResource) DeepCopyInto(out *ItemModuleResource) {
	*out = *in
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]ServiceTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]ConfigMapTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]SecretTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ItemModuleResource.
func (in *ItemModuleResource) DeepCopy() *ItemModuleResource {
	if in == nil {
		return nil
	}
	out := new(ItemModuleResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ItemStatus) DeepCopyInto(out *ItemStatus) {
	*out = *in
	if in.RunningJobNum != nil {
		in, out := &in.RunningJobNum, &out.RunningJobNum
		*out = new(int32)
		**out = **in
	}
	if in.CompletedJobNum != nil {
		in, out := &in.CompletedJobNum, &out.CompletedJobNum
		*out = new(int32)
		**out = **in
	}
	if in.FailedJobNum != nil {
		in, out := &in.FailedJobNum, &out.FailedJobNum
		*out = new(int32)
		**out = **in
	}
	if in.K8sJobStatus != nil {
		in, out := &in.K8sJobStatus, &out.K8sJobStatus
		*out = make(map[string]batchv1alpha1.JobState, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.VolcanoJobStatus != nil {
		in, out := &in.VolcanoJobStatus, &out.VolcanoJobStatus
		*out = make(map[string]batchv1alpha1.JobState, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.ServiceStatus != nil {
		in, out := &in.ServiceStatus, &out.ServiceStatus
		*out = make(map[string]RegularModuleStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.ConfigMapStatus != nil {
		in, out := &in.ConfigMapStatus, &out.ConfigMapStatus
		*out = make(map[string]RegularModuleStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.SecretStatus != nil {
		in, out := &in.SecretStatus, &out.SecretStatus
		*out = make(map[string]RegularModuleStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ItemStatus.
func (in *ItemStatus) DeepCopy() *ItemStatus {
	if in == nil {
		return nil
	}
	out := new(ItemStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Job) DeepCopyInto(out *Job) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Job.
func (in *Job) DeepCopy() *Job {
	if in == nil {
		return nil
	}
	out := new(Job)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Job) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobBatch) DeepCopyInto(out *JobBatch) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobBatch.
func (in *JobBatch) DeepCopy() *JobBatch {
	if in == nil {
		return nil
	}
	out := new(JobBatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JobBatch) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobBatchList) DeepCopyInto(out *JobBatchList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]JobBatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobBatchList.
func (in *JobBatchList) DeepCopy() *JobBatchList {
	if in == nil {
		return nil
	}
	out := new(JobBatchList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JobBatchList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobBatchSpec) DeepCopyInto(out *JobBatchSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobBatchSpec.
func (in *JobBatchSpec) DeepCopy() *JobBatchSpec {
	if in == nil {
		return nil
	}
	out := new(JobBatchSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobBatchStatus) DeepCopyInto(out *JobBatchStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobBatchStatus.
func (in *JobBatchStatus) DeepCopy() *JobBatchStatus {
	if in == nil {
		return nil
	}
	out := new(JobBatchStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobList) DeepCopyInto(out *JobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Job, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobList.
func (in *JobList) DeepCopy() *JobList {
	if in == nil {
		return nil
	}
	out := new(JobList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSpec) DeepCopyInto(out *JobSpec) {
	*out = *in
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Item, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSpec.
func (in *JobSpec) DeepCopy() *JobSpec {
	if in == nil {
		return nil
	}
	out := new(JobSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobState) DeepCopyInto(out *JobState) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobState.
func (in *JobState) DeepCopy() *JobState {
	if in == nil {
		return nil
	}
	out := new(JobState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobStatus) DeepCopyInto(out *JobStatus) {
	*out = *in
	in.State.DeepCopyInto(&out.State)
	if in.ItemStatus != nil {
		in, out := &in.ItemStatus, &out.ItemStatus
		*out = make(map[string]ItemStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobStatus.
func (in *JobStatus) DeepCopy() *JobStatus {
	if in == nil {
		return nil
	}
	out := new(JobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobTemplateSpec) DeepCopyInto(out *JobTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobTemplateSpec.
func (in *JobTemplateSpec) DeepCopy() *JobTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(JobTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegularModuleStatus) DeepCopyInto(out *RegularModuleStatus) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegularModuleStatus.
func (in *RegularModuleStatus) DeepCopy() *RegularModuleStatus {
	if in == nil {
		return nil
	}
	out := new(RegularModuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplate) DeepCopyInto(out *SecretTemplate) {
	*out = *in
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
	in.Secret.DeepCopyInto(&out.Secret)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplate.
func (in *SecretTemplate) DeepCopy() *SecretTemplate {
	if in == nil {
		return nil
	}
	out := new(SecretTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceTemplate) DeepCopyInto(out *ServiceTemplate) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceTemplate.
func (in *ServiceTemplate) DeepCopy() *ServiceTemplate {
	if in == nil {
		return nil
	}
	out := new(ServiceTemplate)
	in.DeepCopyInto(out)
	return out
}