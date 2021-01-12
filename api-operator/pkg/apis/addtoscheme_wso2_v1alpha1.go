package apis

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha2.SchemeBuilder.AddToScheme)
}
