// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
