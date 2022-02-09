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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModuleExecStatusSpec defines the desired state of ModuleExecStatus
type ModuleExecStatusSpec struct {
}

// ModuleExecStatusStatus defines the observed state of ModuleExecStatus
type ModuleExecStatusStatus struct {
	// +optional
	Watchers []Watcher `json:"watchers,omitempty"`
}

type Watcher struct {
	Kind     metav1.GroupVersionKind `json:"kind"`
	Matchers []Matcher               `json:"matchers"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ModuleExecStatus is the Schema for the moduleexecstatuses API
type ModuleExecStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleExecStatusSpec   `json:"spec,omitempty"`
	Status ModuleExecStatusStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModuleExecStatusList contains a list of ModuleExecStatus
type ModuleExecStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModuleExecStatus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModuleExecStatus{}, &ModuleExecStatusList{})
}
