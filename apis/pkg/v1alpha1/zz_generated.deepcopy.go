// +build !ignore_autogenerated

/*
Copyright 2021.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1 "kmodules.xyz/client-go/api/v1"
	metav1alpha1 "kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Action) DeepCopyInto(out *Action) {
	*out = *in
	out.ChartRef = in.ChartRef
	if in.ValuesPatch != nil {
		in, out := &in.ValuesPatch, &out.ValuesPatch
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	if in.ValueOverrides != nil {
		in, out := &in.ValueOverrides, &out.ValueOverrides
		*out = make([]LoadValue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.ReadinessCriteria.DeepCopyInto(&out.ReadinessCriteria)
	if in.Prerequisites != nil {
		in, out := &in.Prerequisites, &out.Prerequisites
		*out = new(Prerequisites)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Action.
func (in *Action) DeepCopy() *Action {
	if in == nil {
		return nil
	}
	out := new(Action)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartRef) DeepCopyInto(out *ChartRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartRef.
func (in *ChartRef) DeepCopy() *ChartRef {
	if in == nil {
		return nil
	}
	out := new(ChartRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KV) DeepCopyInto(out *KV) {
	*out = *in
	out.ValueRef = in.ValueRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KV.
func (in *KV) DeepCopy() *KV {
	if in == nil {
		return nil
	}
	out := new(KV)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadValue) DeepCopyInto(out *LoadValue) {
	*out = *in
	if in.ObjRef != nil {
		in, out := &in.ObjRef, &out.ObjRef
		*out = new(ObjectLocator)
		(*in).DeepCopyInto(*out)
	}
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]KV, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadValue.
func (in *LoadValue) DeepCopy() *LoadValue {
	if in == nil {
		return nil
	}
	out := new(LoadValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Module) DeepCopyInto(out *Module) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Module.
func (in *Module) DeepCopy() *Module {
	if in == nil {
		return nil
	}
	out := new(Module)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Module) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleList) DeepCopyInto(out *ModuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Module, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleList.
func (in *ModuleList) DeepCopy() *ModuleList {
	if in == nil {
		return nil
	}
	out := new(ModuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleSpec) DeepCopyInto(out *ModuleSpec) {
	*out = *in
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]Action, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EdgeList != nil {
		in, out := &in.EdgeList, &out.EdgeList
		*out = make([]metav1alpha1.NamedEdge, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleSpec.
func (in *ModuleSpec) DeepCopy() *ModuleSpec {
	if in == nil {
		return nil
	}
	out := new(ModuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleStatus) DeepCopyInto(out *ModuleStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]apiv1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleStatus.
func (in *ModuleStatus) DeepCopy() *ModuleStatus {
	if in == nil {
		return nil
	}
	out := new(ModuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectLocator) DeepCopyInto(out *ObjectLocator) {
	*out = *in
	in.Src.DeepCopyInto(&out.Src)
	if in.Paths != nil {
		in, out := &in.Paths, &out.Paths
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectLocator.
func (in *ObjectLocator) DeepCopy() *ObjectLocator {
	if in == nil {
		return nil
	}
	out := new(ObjectLocator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectRef) DeepCopyInto(out *ObjectRef) {
	*out = *in
	out.Target = in.Target
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectRef.
func (in *ObjectRef) DeepCopy() *ObjectRef {
	if in == nil {
		return nil
	}
	out := new(ObjectRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Prerequisites) DeepCopyInto(out *Prerequisites) {
	*out = *in
	if in.RequiredResources != nil {
		in, out := &in.RequiredResources, &out.RequiredResources
		*out = make([]v1.GroupVersionResource, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Prerequisites.
func (in *Prerequisites) DeepCopy() *Prerequisites {
	if in == nil {
		return nil
	}
	out := new(Prerequisites)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReadinessCriteria) DeepCopyInto(out *ReadinessCriteria) {
	*out = *in
	out.Timeout = in.Timeout
	if in.ResourcesExist != nil {
		in, out := &in.ResourcesExist, &out.ResourcesExist
		*out = make([]v1.GroupVersionResource, len(*in))
		copy(*out, *in)
	}
	if in.WaitFors != nil {
		in, out := &in.WaitFors, &out.WaitFors
		*out = make([]WaitFlags, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReadinessCriteria.
func (in *ReadinessCriteria) DeepCopy() *ReadinessCriteria {
	if in == nil {
		return nil
	}
	out := new(ReadinessCriteria)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueRef) DeepCopyInto(out *ValueRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueRef.
func (in *ValueRef) DeepCopy() *ValueRef {
	if in == nil {
		return nil
	}
	out := new(ValueRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WaitFlags) DeepCopyInto(out *WaitFlags) {
	*out = *in
	out.Resource = in.Resource
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	out.Timeout = in.Timeout
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WaitFlags.
func (in *WaitFlags) DeepCopy() *WaitFlags {
	if in == nil {
		return nil
	}
	out := new(WaitFlags)
	in.DeepCopyInto(out)
	return out
}
