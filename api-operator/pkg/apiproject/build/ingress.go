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

package build

import (
	"context"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject/names"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/tls"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/swagger"
	"k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

// FromIngress builds build.ProjectsMap with actions needed updated in Router from given ingresses
func FromIngress(ctx context.Context, reqInfo *common.RequestInfo, ingresses []*ingress.Ingress, projectsToBeUpdated, existingProjects apiproject.ProjectSet) (*ProjectsMap, error) {
	projectMap := ProjectsMap{}
	// Initialize project action with type delete
	for p := range projectsToBeUpdated {
		projectMap[p] = &Project{
			// Set default action to Delete, so if ingress not define a particular existing project it is deleted
			Action: Delete,
			// Default OAS project
			OAS: defaultOpenAPI(p),
			// Keep TlsSecret as nil to check whether, TlsSecret is set by ingress
			TlsCertificate: nil,
		}
	}

	for _, ing := range ingresses {
		if k8s.IsDeleted(ing) {
			// Ingress is deleted, no need to process
			continue
		}

		if err := processDefaultBackend(ctx, reqInfo, projectsToBeUpdated, &projectMap, ing); err != nil {
			return nil, err
		}
		if err := processIngressRules(ctx, reqInfo, projectsToBeUpdated, &projectMap, ing); err != nil {
			return nil, err
		}
		if err := processIngressTls(ctx, reqInfo, projectsToBeUpdated, &projectMap, ing); err != nil {
			return nil, err
		}
	}

	// Already not existing projects can not be deleted
	// Do nothing for those projects
	for project := range projectsToBeUpdated {
		if projectMap[project].Action == Delete && !existingProjects[project] {
			projectMap[project].Action = DoNothing
		}
	}

	return &projectMap, nil
}

// processDefaultBackend go through ingress default backend (if default project should be updated) and updates the
// Open API Spec (openapi3.Swagger) of names.DefaultBackendProject in the ProjectsMap and the action to ForceUpdate
func processDefaultBackend(ctx context.Context, reqInfo *common.RequestInfo, projects apiproject.ProjectSet, projectMap *ProjectsMap, ing *ingress.Ingress) error {
	log := reqInfo.Log
	pMap := *projectMap
	// Default backend
	// Check whether the ingress contributes to the default backend project
	if projects[names.DefaultBackendProject] && ing.Spec.Backend != nil {
		if pMap[names.DefaultBackendProject].OAS.Servers != nil {
			// skip this default backend configuration
			// give priority to older ingress
			log.Info("Skipping the default backend configuration, since it is already defined by old ingress",
				"new_ingress", ing)
		} else {
			// Do not validate service and port existence since user may add ingress first and service later
			u := urlFromIngBackend(ing, ing.Spec.Backend)
			pMap[names.DefaultBackendProject].OAS.Servers = oasServers(u)

			certs, err := getBackendCerts(ctx, reqInfo, ing)
			if err != nil {
				log.Error(err, "Could not get backend certs and skipping the default backend configuration in this ingress",
					"ingress", ing)
			}
			pMap[names.DefaultBackendProject].BackendCertificates = certs
			pMap[names.DefaultBackendProject].Action = ForceUpdate
		}
	}

	return nil
}

// processIngressRules go through ingress rules and updates the Open API Spec (openapi3.Swagger) in
// the ProjectsMap and the action to ForceUpdate
func processIngressRules(ctx context.Context, reqInfo *common.RequestInfo, projects apiproject.ProjectSet, projectMap *ProjectsMap, ing *ingress.Ingress) error {
	log := reqInfo.Log
	pMap := *projectMap

	for _, rule := range ing.Spec.Rules {
		pj := names.HostToProject(rule.Host)
		validPj := false

		// check whether the ingress contributes to the project
		if projects[pj] {
			for _, path := range rule.HTTP.Paths {
				oasPath := path.Path
				if *path.PathType == v1beta1.PathTypeExact {
					if strings.HasSuffix(oasPath, "/*") {
						// check debug level
						// TODO: (renuka) should this be skipped or corrected with removing suffix or treat as prefixed type
						log.Info("Skipping the path configuration for the host defined, since path type is \"exact\" and path is suffixed with \"*\"",
							"ingress", ing, "host", rule.Host, "path", oasPath)
						oasPath += "/*"
					}
				} else {
					// path type is Prefix or ImplementationSpecific
					// double check for the existence of suffix
					if !strings.HasSuffix(oasPath, "/*") {
						oasPath += "/*"
					}
				}

				if pMap[pj].OAS.Paths[oasPath] != nil {
					// skip this path
					// give priority to older ingress
					log.Info("Skipping the path configuration for the host defined, since it is already defined by old ingress",
						"new_ingress", ing, "host", rule.Host, "path", oasPath)
					continue
				}

				// Do not validate service and port existence since user may add ingress first and service later
				u := urlFromIngBackend(ing, &path.Backend)
				pMap[pj].OAS.Paths[oasPath] = oasPathItem(u)
				validPj = true
			}
			// Update the project if valid
			if validPj {
				certs, err := getBackendCerts(ctx, reqInfo, ing)
				if err != nil {
					log.Error(err, "Could not get backend certs in the ingress",
						"ingress", ing, "ingress_project", pj)
				}
				pMap[pj].BackendCertificates = certs
				pMap[pj].Action = ForceUpdate
			}
		}
	}

	return nil
}

// processIngressTls go through ingress TLS rules and updates the Project.TlsCertificate in
// the ProjectsMap and the action to ForceUpdate
func processIngressTls(ctx context.Context, reqInfo *common.RequestInfo, projects apiproject.ProjectSet, projectMap *ProjectsMap, ing *ingress.Ingress) error {
	log := reqInfo.Log
	pMap := *projectMap

	for _, ingTls := range ing.Spec.TLS {
		// Check secret
		secret := &v1.Secret{}
		if ingTls.SecretName != "" {
			if err := reqInfo.Client.Get(ctx, types.NamespacedName{Namespace: ing.Namespace, Name: ingTls.SecretName}, secret); err != nil {
				if k8serrors.IsNotFound(err) {
					log.Error(err, "TLS secret not found and skipping TLS configuration for the hosts defined in the ingress", "ingress", ing, "hosts", ingTls.Hosts)
					// continue with other tls configs in the ingress with only skipping this config
					continue
				}
				return err
			}
		}

		for _, host := range ingTls.Hosts {
			pj := names.HostToProject(host)

			// check whether the ingress contributes to the project
			if projects[pj] {
				if pMap[pj].TlsCertificate != nil {
					// skip this tls configuration
					// give priority to older ingress
					log.Info("Skipping the TLS configuration for the host defined, since it is already defined by old ingress",
						"new_ingress", ing, "host", host)
					continue
				}

				tlsCertificate, err := tlsCertFromIngTls(secret, ing)
				if err != nil {
					log.Error(err, "Invalid tls secret and skipping TLS configuration for the host defined in the ingress", "ingress", ing, "host", host)
					continue
				}
				pMap[pj].TlsCertificate = tlsCertificate
				// Do not make an Action as ForceUpdate since project should have
				// HTTP rules or default backend to have a project
			}
		}
	}
	return nil
}

// getBackendCerts returns list of backend certificates in the given ingress
func getBackendCerts(ctx context.Context, reqInfo *common.RequestInfo, ing *ingress.Ingress) ([]*TlsSecret, error) {
	log := reqInfo.Log
	var certs []*TlsSecret

	for _, name := range ing.ParsedAnnotations.Tls.BackendCerts {
		// Check secret
		secret := &v1.Secret{}
		if err := reqInfo.Client.Get(ctx, types.NamespacedName{Namespace: ing.Namespace, Name: name}, secret); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Error(err, "Backend secret not found and skipping it", "ingress", ing)
				// continue with other secrets
				continue
			}
			return nil, err
		}

		caCert, ok := secret.Data["ca.crt"]
		if !ok {
			log.Error(nil, "Invalid tls secret and skipping backend cert in the secret", "ingress", ing, "secret", name)
			// continue with other secrets
			continue
		}
		certs = append(certs, &TlsSecret{TrustedCa: caCert})
	}
	return certs, nil
}

func tlsCertFromIngTls(secret *v1.Secret, ing *ingress.Ingress) (*TlsSecret, error) {
	tlsCert := TlsSecret{}

	crt, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("tls certificate not found in the field \"tls.crt\" of secret %s", secret.String()))
	}
	tlsCert.CertificateChain = crt

	key, ok := secret.Data["tls.key"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("tls key not found in the field \"tls.key\" of secret %s", secret.String()))
	}
	tlsCert.PrivateKey = key

	if ing.ParsedAnnotations.Tls.Mode == tls.Mutual {
		caCert, ok := secret.Data["ca.crt"]
		if !ok {
			return nil, errors.New(fmt.Sprintf("tls ca cert not found in the field \"ca.crt\" of secret %s", secret.String()))
		}
		tlsCert.TrustedCa = caCert
	}

	return &tlsCert, nil
}

func isPortExists(svc *v1.Service, port int) bool {
	for _, ports := range svc.Spec.Ports {
		if ports.Port == int32(port) {
			return true
		}
	}
	return false
}

func urlFromIngBackend(ing *ingress.Ingress, backend *v1beta1.IngressBackend) string {
	protocol := "http"
	if ing.ParsedAnnotations.Tls.BackendProtocol == tls.HTTPS || ing.ParsedAnnotations.Tls.Mode == tls.Passthrough {
		protocol = "https"
	}
	// Using only backend.ServiceName
	return fmt.Sprintf("%s://%s.%s:%s", protocol, ing.Namespace, backend.ServiceName, backend.ServicePort.String())
}

func oasPathItem(url string) *openapi3.PathItem {
	return &openapi3.PathItem{
		Summary:     "",
		Description: "",
		Servers:     oasServers(url),
	}
}

func oasServers(url string) openapi3.Servers {
	return openapi3.Servers{{
		URL:         url,
		Description: "",
	}}
}

func defaultOpenAPI(projectName string) *openapi3.Swagger {
	return &openapi3.Swagger{
		ExtensionProps: openapi3.ExtensionProps{
			Extensions: map[string]interface{}{
				swagger.VhostExtension: names.ProjectToHost(projectName),
				swagger.SpecExtension:  "ingress",
			},
		},
		OpenAPI: "3.0.0",
		Info: openapi3.Info{
			Title:       projectName,
			Description: projectName,
			Version:     "v1",
		},
		Servers:      nil,
		Paths:        openapi3.Paths{},
		Components:   openapi3.Components{},
		Security:     nil,
		ExternalDocs: nil,
	}
}
