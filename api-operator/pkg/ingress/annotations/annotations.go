package annotations

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/tls"
	"k8s.io/api/networking/v1beta1"
)

type Ingress struct {
	Tls tls.Config
}

// TODO (renuka) handle errors
func ParseIngress(ing *v1beta1.Ingress) Ingress {
	return Ingress{Tls: tls.Parse(ing)}
}
