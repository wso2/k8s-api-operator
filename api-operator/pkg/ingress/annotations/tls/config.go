package tls

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/parser"
	"k8s.io/api/networking/v1beta1"
	"strings"
)

type Mode string

const (
	NoTls       = Mode("no-tls")
	Simple      = Mode("simple")
	Mutual      = Mode("mtls")
	Passthrough = Mode("passthrough")
	Origination = Mode("origination")
)

// Annotations
const (
	tlsMode = "tls-mode"
)

type Config struct {
	TlsMode               Mode
	TlsOriginationEnabled bool
	// TODO (renuka)
	//TlsOriginationCerts
}

func Parse(ing *v1beta1.Ingress) Config {
	return Config{
		TlsMode: Mode(strings.ToLower(ing.Annotations[parser.GetAnnotationWithPrefix(tlsMode)])),
	}
}
