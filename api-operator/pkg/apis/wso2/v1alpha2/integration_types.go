/*
 * Copyright (c) 2020 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IntegrationSpec defines the desired state of Integration
type IntegrationSpec struct {
	// Docker image consist of micro integrator runtime and synapse configs
	Image string `json:"image"`
	// Specification related to deployment
	DeploySpec DeploySpec `json:"deploySpec,omitempty"`
	// Auto scale spec
	AutoScale AutoScale `json:"autoScale,omitempty"`
	// Ports to expose
	Expose Expose `json:"expose,omitempty"`
	// Docker image credentials if the Image is in private registry
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
	// List of environment variables to set for the integration.
	Env []corev1.EnvVar `json:"env,omitempty"`
	// List of environment variable references set for the integration.
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

// DeploySpec contains properties related to deployment
type DeploySpec struct {
	// Initial minimum number of replicas in Integration
	// Default value "<empty>".
	// +optional
	MinReplicas int32 `json:"minReplicas,omitempty"`

	ConfigMapDetails []ConfigMapDetails `json:"configmapDetails,omitempty"`

	// Cpu request of containers in the pod
	// Default value "<empty>".
	// +optional
	ReqCpu string `json:"requestCPU,omitempty"`
	// Memory request of containers in the pod
	// Default value "<empty>".
	// +optional
	ReqMemory string `json:"reqMemory,omitempty"`
	// CPU limit of containers in the pod
	// Default value "<empty>".
	// +optional
	LimitCpu string `json:"cpuLimit,omitempty"`
	// Memory limit of containers in the pod
	// Default value "<empty>".
	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// LivenessProbe for containers in the pod
	// Default value "<empty>".
	// +optional
	LivenessProbe corev1.Probe `json:"livenessProbe,omitempty"`
	// ReadinessProbe for containers in the pod
	// Default value "<empty>".
	// +optional
	ReadinessProbe corev1.Probe `json:"readinessProbe,omitempty"`

	PullPolicy PullPolicy `json:"pullPolicy,omitempty"`
}

// AutoScale defines the properties related to Auto scaling of pods
type AutoScale struct {
	// Defines if auto scaling needs to be enabled
	// Default value "<empty>".
	// +optional
	Enabled string `json:"enabled,omitempty"`
	// Defines maximum number of replicas of the Integration deployment
	// Default value "<empty>".
	// +optional
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
}
type PullPolicy string

const (
	// PullAlways means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

type ConfigMapDetails struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
	FileName  string `json:"fileName,omitempty"`
}

// Expose defines the ports needs to be exposed
type Expose struct {
	// PassthroPort HTTP/HTTPs traffic serving ports
	PassthroPort int32 `json:"passthroPort,omitempty"`
	// InboundPorts any traffic serving port that needs to be exposed
	InboundPorts []int32 `json:"inboundPorts,omitempty"`
}

// IntegrationStatus defines the observed state of Integration
type IntegrationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	// Name of the created service in the Integration deployment
	ServiceName string `json:"serviceName"`
	// Status of the Integration deployment
	Readiness string `json:"readiness"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Integration is the Schema for the integrations API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=integrations,scope=Namespaced
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.readiness`
// +kubebuilder:printcolumn:name="Service-Name",type=string,JSONPath=`.status.serviceName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Integration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationSpec   `json:"spec,omitempty"`
	Status IntegrationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IntegrationList contains a list of Integration
type IntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Integration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Integration{}, &IntegrationList{})
}
