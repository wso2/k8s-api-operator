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
//   Action: ForceUpdate
//   OAS: swagger.yaml
// example2_com:
//   Action: Delete
//   OAS: empty_swagger.yaml
//
type ProjectsMap map[string]*Project

// Project represents action to be done to the envoy microgateway
type Project struct {
	// Action of the action
	Action action
	// OAS definition to be updated
	OAS *openapi3.Swagger
	// TlsCertificate of the project for TLS termination
	TlsCertificate *TlsCertificate
	// BackendCertificates of the backends for TLS origination
	BackendCertificates []*TlsCertificate
}

type TlsCertificate struct {
	CertificateChain []byte
	PrivateKey       []byte
	TrustedCa        []byte
}

func (c *TlsCertificate) String() string {
	if c == nil {
		return "{}"
	}

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

// action represents the type of action
type action string

// Project types
const (
	// Delete action defines an existing project should be deleted
	Delete = action("Delete")
	// ForceUpdate action defines an existing project should be updated or created if not exists
	ForceUpdate = action("ForceUpdate")
	// DoNothing action defines a new project that are invalid should be ignored
	DoNothing = action("DoNothing")
)
