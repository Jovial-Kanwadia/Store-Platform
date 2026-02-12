/*
Copyright 2026.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StoreSpec defines the desired state of Store
type StoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Engine type: woo | medusa
	Engine string `json:"engine"`

	// Plan or size (small, medium, etc)
	Plan string `json:"plan"`
}

// StoreStatus defines the observed state of Store
type StoreStatus struct {
	// Phase is the current lifecycle phase (Provisioning, Ready, Failed)
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the last generation of the Store that was successfully reconciled
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// URL is the external endpoint for the store
	URL string `json:"url,omitempty"`

	// Message is a human-readable description of the current state
	// +optional
	Message string `json:"message,omitempty"`

	// Reason is a machine-readable reason code for the current phase
	// +optional
	Reason string `json:"reason,omitempty"`

	// Conditions store the detailed state history
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Store is the Schema for the stores API
type Store struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Store
	// +required
	Spec StoreSpec `json:"spec"`

	// status defines the observed state of Store
	// +optional
	Status StoreStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// StoreList contains a list of Store
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Store `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Store{}, &StoreList{})
}
