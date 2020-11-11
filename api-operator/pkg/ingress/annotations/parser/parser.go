package parser

import (
	"fmt"
	"k8s.io/api/networking/v1beta1"
)

const (
	// DefaultPrefix defines the default annotation prefix used in the WSO2 microgateway ingress controller.
	DefaultPrefix = "microgateway.ingress.wso2.com"

	TlsMode       = "tls-mode"
	ApiManagement = "api-management"
)

var (
	// Prefix defines the annotation prefix which is mutable
	Prefix = DefaultPrefix
)

type Parser interface {
	Parse(*v1beta1.Ingress)
}

func GetAnnotationWithPrefix(name string) string {
	return fmt.Sprintf("%v/%v", Prefix, name)
}
