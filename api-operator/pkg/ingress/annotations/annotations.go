package annotations

import "fmt"

const (
	// DefaultPrefix defines the default annotation prefix used in the WSO2 microgateway ingress controller.
	DefaultPrefix = "microgateway.ingress.wso2.com"
)

var (
	// Prefix defines the annotation prefix which is mutable
	Prefix = DefaultPrefix
)

func GetAnnotation(name string) string {
	return fmt.Sprintf("%v/%v", Prefix, name)
}
