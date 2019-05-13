package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TargetEndpointSpec defines the desired state of TargetEndpoint
// +k8s:openapi-gen=true
type TargetEndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Type             string           `json:"type"`
	Protocol         string           `json:"protocol"`
	Hostname         string           `json:"hostname"`
	Port             int32            `json:"port"`
	Deploy           Deploy           `json:"deploy"`
	EndpointName     string           `json:"endpointName"`
	EndpointSecurity EndpointSecurity `json:"endpointSecurity"`
}

// TargetEndpointStatus defines the observed state of TargetEndpoint
// +k8s:openapi-gen=true
type TargetEndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

type EndpointSecurity struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type Deploy struct {
	Name        string `json:"name"`
	DockerImage string `json:"dockerImage"`
	Count       int32  `json:"count"`
}
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TargetEndpoint is the Schema for the targetendpoints API
// +k8s:openapi-gen=true
type TargetEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetEndpointSpec   `json:"spec,omitempty"`
	Status TargetEndpointStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TargetEndpointList contains a list of TargetEndpoint
type TargetEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TargetEndpoint{}, &TargetEndpointList{})
}
