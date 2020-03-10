package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceSpec defines the desired state of Service
// +k8s:openapi-gen=true
type ServiceSpec struct {
	// DeprecatedGeneration was used prior in Kubernetes versions <1.11
	// when metadata.generation was not being incremented by the api server
	//
	// This property will be dropped in future Knative releases and should
	// not be used - use metadata.generation
	//
	// Tracking issue: https://github.com/knative/serving/issues/643
	//
	// +optional
	DeprecatedGeneration int64 `json:"generation,omitempty"`

	// DeprecatedRunLatest defines a simple Service. It will automatically
	// configure a route that keeps the latest ready revision
	// from the supplied configuration running.
	// +optional
	DeprecatedRunLatest *RunLatestType `json:"runLatest,omitempty"`

	// DeprecatedPinned is DEPRECATED in favor of ReleaseType
	// +optional
	DeprecatedPinned *PinnedType `json:"pinned,omitempty"`

	// DeprecatedManual mode enables users to start managing the underlying Route and Configuration
	// resources directly.  This advanced usage is intended as a path for users to graduate
	// from the limited capabilities of Service to the full power of Route.
	// +optional
	DeprecatedManual *ManualType `json:"manual,omitempty"`

	// Release enables gradual promotion of new revisions by allowing traffic
	// to be split between two revisions. This type replaces the deprecated Pinned type.
	// +optional
	DeprecatedRelease *ReleaseType `json:"release,omitempty"`

	// We are moving to a shape where the Configuration and Route specifications
	// are inlined into the Service, which gives them compatible shapes.  We are
	// staging this change here as a path to this in v1beta1, which drops the
	// "mode" based specifications above.  Ultimately all non-v1beta1 fields will
	// be deprecated, and then dropped in v1beta1.
	ConfigurationSpec `json:",inline"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Configuration represents the "floating HEAD" of a linear history of Revisions,
// and optionally how the containers those revisions reference are built.
// Users create new Revisions by updating the Configuration's spec.
// The "latest created" revision's name is available under status, as is the
// "latest ready" revision's name.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Configuration (from the client).
	// +optional
	Spec ConfigurationSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the Configuration (from the controller).
	// +optional
	Status ConfigurationStatus `json:"status,omitempty"`
}
type ConfigurationSpec struct {
	// DeprecatedGeneration was used prior in Kubernetes versions <1.11
	// when metadata.generation was not being incremented by the api server
	//
	// This property will be dropped in future Knative releases and should
	// not be used - use metadata.generation
	//
	// Tracking issue: https://github.com/knative/serving/issues/643
	//
	// +optional
	DeprecatedGeneration int64 `json:"generation,omitempty"`

	// DeprecatedRevisionTemplate holds the latest specification for the Revision to
	// be stamped out. If a Build specification is provided, then the
	// DeprecatedRevisionTemplate's BuildName field will be populated with the name of
	// the Build object created to produce the container for the Revision.
	// DEPRECATED Use Template instead.
	// +optional

	// Template holds the latest specification for the Revision to
	// be stamped out.
	// +optional
	Template RevisionTemplateSpec `json:"template,omitempty"`
}

type RevisionTemplateSpec struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec RevisionSpec `json:"spec,omitempty"`
}

type RevisionSpec struct {
	corev1.PodSpec `json:",inline"`
}


// ConfigurationStatusFields holds all of the non-duckv1beta1.Status status fields of a Route.
// These are defined outline so that we can also inline them into Service, and more easily
// copy them.
type ConfigurationStatusFields struct {
	// LatestReadyRevisionName holds the name of the latest Revision stamped out
	// from this Configuration that has had its "Ready" condition become "True".
	// +optional
	LatestReadyRevisionName string `json:"latestReadyRevisionName,omitempty"`

	// LatestCreatedRevisionName is the last revision that was created from this
	// Configuration. It might not be ready yet, for that use LatestReadyRevisionName.
	// +optional
	LatestCreatedRevisionName string `json:"latestCreatedRevisionName,omitempty"`
}

// ConfigurationStatus communicates the observed state of the Configuration (from the controller).
type ConfigurationStatus struct {
	ConfigurationStatusFields `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigurationList is a list of Configuration resources
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Configuration `json:"items"`
}

// ManualType contains the options for configuring a manual service. See ServiceSpec for
// more details.
type ManualType struct {
	// DeprecatedManual type does not contain a configuration as this type provides the
	// user complete control over the configuration and route.
}

// ReleaseType contains the options for slowly releasing revisions. See ServiceSpec for
// more details.
type ReleaseType struct {
	// Revisions is an ordered list of 1 or 2 revisions. The first will
	// have a TrafficTarget with a name of "current" and the second will have
	// a name of "candidate".
	// +optional
	Revisions []string `json:"revisions,omitempty"`

	// RolloutPercent is the percent of traffic that should be sent to the "candidate"
	// revision. Valid values are between 0 and 99 inclusive.
	// +optional
	RolloutPercent int `json:"rolloutPercent,omitempty"`

	// The configuration for this service. All revisions from this service must
	// come from a single configuration.
	// +optional
	Configuration ConfigurationSpec `json:"configuration,omitempty"`
}

// ReleaseLatestRevisionKeyword is a shortcut for usage in the `release` mode
// to refer to the latest created revision.
// See #2819 for details.
const ReleaseLatestRevisionKeyword = "@latest"

// RunLatestType contains the options for always having a route to the latest configuration. See
// ServiceSpec for more details.
type RunLatestType struct {
	// The configuration for this service.
	// +optional
	Configuration ConfigurationSpec `json:"configuration,omitempty"`
}

// PinnedType is DEPRECATED. ReleaseType should be used instead. To get the behavior of PinnedType set
// ReleaseType.Revisions to []string{PinnedType.RevisionName} and ReleaseType.RolloutPercent to 0.
type PinnedType struct {
	// The revision name to pin this service to until changed
	// to a different service type.
	// +optional
	RevisionName string `json:"revisionName,omitempty"`

	// The configuration for this service.
	// +optional
	Configuration ConfigurationSpec `json:"configuration,omitempty"`
}



// ServiceStatus defines the observed state of Service
// +k8s:openapi-gen=true
type ServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Service is the Schema for the services API
// +k8s:openapi-gen=true
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
