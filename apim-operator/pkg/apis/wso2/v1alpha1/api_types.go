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
	Name                     string              `json:"name"`
	Context                  string              `json:"context"`
	Version                  string              `json:"version"`
	Description              string              `json:"description"`
	Tags                     []string            `json:"tags"`
	Endpoints                []Endpoint          `json:"endpoints"`
	RequestInterceptor       string              `json:"requestInterceptor"`
	ResponseInterceptor      string              `json:"responseInterceptor"`
	AuthorizationHeader      string              `json:"authorizationHeader"`
	Labels                   []string            `json:"labels"`
	URLPatterns              []URLPattern        `json:"urlPatterns"`
	Security                 []string            `json:"security"`
	SubscriptionTiers        []string            `json:"subscriptionTiers"`
	AdvancedThrottlingPolicy string              `json:"advancedThrottlingPolicy"`
	BusinessInformation      BusinessInformation `json:"businessInformation"`
	APIProperties            []APIProperty       `json:"apiProperties"`
	Mode                     Mode                `json:"mode"`
	ReplicaCount             int32               `json:"replicaCount"`
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

type Endpoint struct {
	Type         string `json:"type"`
	Protocol     string `json:"protocol"`
	Hostname     string `json:"hostname"`
	Port         int32  `json:"port"`
	DockerImage  string `json:"dockerImage"`
	EndpointName string `json:"endpointName"`
}

type URLPattern struct {
	Path                     string     `json:"path"`
	Method                   string     `json:"method"`
	Scopes                   []string   `json:"scopes"`
	RequestInterceptor       string     `json:"requestInterceptor"`
	ResponseInterceptor      string     `json:"responseInterceptor"`
	Endpoints                []Endpoint `json:"endpoints"`
	AdvancedThrottlingPolicy string     `json:"advancedThrottlingPolicy"`
}

type BusinessInformation struct {
	BusinessOwner       string `json:"businessOwner"`
	BusinessOwnerEmail  string `json:"businessOwnerEmail"`
	TechnicalOwner      string `json:"technicalOwner"`
	TechnicalOwnerEmail string `json:"technicalOwnerEmail"`
}

type APIProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Mode string

const (
	PrivateJet Mode = "privateJet"
	Sidecar    Mode = "sidecar"
	Shared     Mode = "shared"
)

func init() {
	SchemeBuilder.Register(&API{}, &APIList{})
}
