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

// APISpec defines the desired state of API
// +k8s:openapi-gen=true
type APISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Mode            Mode       `json:"mode,omitempty"`
	UpdateTimeStamp string     `json:"updateTimeStamp,omitempty"`
	Replicas        int        `json:"replicas"`
	Definition      Definition `json:"definition"`
	Override        bool       `json:"override,omitempty"`
	Version         string     `json:"version,omitempty"`
}

// APIStatus defines the observed state of API
// +k8s:openapi-gen=true
type APIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// API is the Schema for the apis API
// +k8s:openapi-gen=true
type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APISpec   `json:"spec,omitempty"`
	Status APIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIList contains a list of API
type APIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []API `json:"items"`
}

//Definition contains api definition related values
type Definition struct {
	SwaggerConfigmapNames []string     `json:"swaggerConfigmapNames"`
	Type                  string       `json:"type,omitempty"`
	Interceptors          Interceptors `json:"interceptors,omitempty"`
}

type Interceptors struct {
	Ballerina []string `json:"ballerina,omitempty"`
	Java      []string `json:"java,omitempty"`
}

type Mode string

const (
	PrivateJet Mode = "privateJet"
	Sidecar    Mode = "sidecar"
	Shared     Mode = "shared"
	Serverless Mode = "serverless"
)

func (c Mode) String() string {
	return string(c)
}

func init() {
	SchemeBuilder.Register(&API{}, &APIList{})
}
