package ingress

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations"
	"k8s.io/api/networking/v1beta1"
)

type Ingress struct {
	v1beta1.Ingress
	ParsedAnnotations annotations.Ingress
}

// TODO (renuka) handle errors
func WithAnnotations(ingress *v1beta1.Ingress) *Ingress {
	return &Ingress{
		Ingress:           *ingress,
		ParsedAnnotations: annotations.ParseIngress(ingress),
	}
}
