package action

import "github.com/getkin/kin-openapi/openapi3"

// ProjectsMap represents an action needed to be sent to the envoy microgateway
// Maps project -> action
//
// example1_com:
//   Type: Update
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
}

// Type represents the type of action
type Type int

// Project types
const (
	Delete = Type(1)
	Update = Type(2)
)
