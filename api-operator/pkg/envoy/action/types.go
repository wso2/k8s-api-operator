package action

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

// ProjectsMap represents an action needed to be sent to the envoy microgateway
// Maps project -> action
//
// example1_com:
//   Type: ForceUpdate
//   OAS: swagger.yaml
// example2_com:
//   Type: Delete
//   OAS: empty_swagger.yaml
//
type ProjectsMap map[string]*Project

// Project represents action to be done to the envoy microgateway
type Project struct {
	// Type of the action
	Type Type
	// OAS definition to be updated
	OAS *openapi3.Swagger
	// TlsCertificate of the project for TLS termination
	TlsCertificate *TlsCertificate
}

type TlsCertificate struct {
	CertificateChain []byte
	PrivateKey       []byte
	TrustedCa        []byte
}

func (c *TlsCertificate) String() string {
	elem := make([]string, 0, 3)
	if c.CertificateChain != nil {
		elem = append(elem, "CertificateChain")
	}
	if c.PrivateKey != nil {
		elem = append(elem, "PrivateKey")
	}
	if c.TrustedCa != nil {
		elem = append(elem, "TrustedCa")
	}

	return fmt.Sprintf("{%s}", strings.Join(elem, ", "))
}

// Type represents the type of action
type Type int

func (t Type) String() string {
	switch t {
	case Delete:
		return "Delete"
	case ForceUpdate:
		return "ForceUpdate"
	case DoNothing:
		return "DoNothing"
	}
	return "Unsupported Action Type"
}

// Project types
const (
	Delete      = Type(1)
	ForceUpdate = Type(2)
	// DoNothing for new projects that are invalid
	DoNothing = Type(3)
)
