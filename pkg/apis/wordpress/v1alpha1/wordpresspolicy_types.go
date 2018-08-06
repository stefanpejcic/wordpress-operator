/*
Copyright 2018 Pressinfra SRL.

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

const (
	WordpressPolicyAnnotationPrefix = "policy.wordpress.presslabs.org/"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WordpressPolicySpec defines the desired state of WordpressPolicy
type WordpressPolicySpec struct {
	// Selector defines the label selector for applying the policy
	// If not specified, applies the policy to all Wordpress resources
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// Priority defines the order in which the policies apply. They apply in
	// reverse priority order
	Priority int32 `json:"priority,omitempty"`
	// Policy template to enforce
	Template WordpressTemplateSpec `json:"template"`
}

type WordpressTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WordpressRuntimeSpec `json:"spec,omitempty"`
}

// WordpressPolicyStatus defines the observed state of WordpressPolicy
type WordpressPolicyStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WordpressPolicy is the Schema for the wordpresspolicies API
// +k8s:openapi-gen=true
type WordpressPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordpressPolicySpec   `json:"spec,omitempty"`
	Status WordpressPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WordpressPolicyList contains a list of WordpressPolicy
type WordpressPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WordpressPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WordpressPolicy{}, &WordpressPolicyList{})
}
