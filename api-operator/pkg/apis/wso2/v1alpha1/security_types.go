// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecuritySpec defines the desired state of Security
// +k8s:openapi-gen=true
type SecuritySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Type           string           `json:"type"`
	SecurityConfig []SecurityConfig `json:"securityConfig"`
}

// SecurityStatus defines the observed state of Security
// +k8s:openapi-gen=true
type SecurityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Security is the Schema for the securities API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="SECURITY_TYPE",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`
type Security struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecuritySpec   `json:"spec,omitempty"`
	Status SecurityStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityList contains a list of Security
type SecurityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Security `json:"items"`
}

type SecurityConfig struct {
	Certificate          string `json:"certificate"`
	Alias                string `json:"alias"`
	Endpoint             string `json:"endpoint"`
	Credentials          string `json:"credentials"`
	Issuer               string `json:"issuer"`
	Audience             string `json:"audience"`
	ValidateSubscription bool   `json:"validateSubscription,omitempty"`
	ValidateAllowedAPIs  bool   `json:"validateAllowedAPIs,omitempty"`
	JwksURL              string `json:"jwksURL"`
}

func init() {
	SchemeBuilder.Register(&Security{}, &SecurityList{})
}
