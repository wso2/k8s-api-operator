package tls

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/parser"
	networking "k8s.io/api/networking/v1beta1"
	"regexp"
	"strings"
)

// TLS modes
const (
	Simple      = mode("simple")
	Mutual      = mode("mtls")
	Passthrough = mode("passthrough")
)

// Backend protocols
const (
	HTTPS = protocol("HTTPS")
	HTTP  = protocol("HTTP")
)

// Annotations suffixes
const (
	modeKey            = "tls-mode"
	backendCertsKey    = "backend-certs"
	backendProtocolKey = "backend-protocol"
)

var (
	modeRegex            = regexp.MustCompile(`^(simple|mtls|passthrough)$`)
	backendProtocolRegex = regexp.MustCompile(`^(HTTP|HTTPS)$`)
)

type Config struct {
	// Mode is the TLS termination mode (or pass-through)
	// Could be one of "simple", "mtls" or "passthrough
	// Default to "simple"
	Mode mode

	// BackendCerts defines the certs for TLS origination to communicate with backend
	// Default to ""
	BackendCerts []string

	// BackendProtocol defines the protocol of backend to communicate with backend
	// Default to "HTTP"
	BackendProtocol protocol
}

type mode string
type protocol string

func Parse(ing *networking.Ingress) Config {
	md, err := parser.GetStringAnnotation(ing, modeKey)
	if err != nil || modeRegex.MatchString(md) {
		md = string(Simple)
	}

	secretStr, _ := parser.GetStringAnnotation(ing, backendCertsKey)
	secrets := strings.Split(secretStr, ",")
	backCerts := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		backCerts = append(backCerts, strings.TrimSpace(secret))
	}

	backProto, err := parser.GetStringAnnotation(ing, backendProtocolKey)
	if err != nil || backendProtocolRegex.MatchString(backProto) {
		backProto = string(HTTP)
	}

	return Config{
		Mode:            mode(md),
		BackendCerts:    backCerts,
		BackendProtocol: protocol(backProto),
	}
}
