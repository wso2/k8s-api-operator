package action

import (
	"context"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/names"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/tls"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

func FromProjects(ctx context.Context, reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress, projects map[string]bool) (*ProjectsMap, error) {
	projectMap := ProjectsMap{}
	// Initialize project action with type delete
	for p := range projects {
		projectMap[p] = &Project{
			Type: Delete,
			OAS:  defaultOpenAPI(p),
			// Keep TlsCertificate as nil to check whether, TlsCertificate is set by ingress
			TlsCertificate: nil,
		}
	}

	for _, ing := range ingresses {
		if k8s.IsDeleted(ing) {
			// Ingress is deleted, no need to process
			continue
		}

		if err := processDefaultBackend(ctx, reqInfo, projects, &projectMap, ing); err != nil {
			return nil, err
		}
		if err := processIngressRules(ctx, reqInfo, projects, &projectMap, ing); err != nil {
			return nil, err
		}
		if err := processIngressTls(ctx, reqInfo, projects, &projectMap, ing); err != nil {
			return nil, err
		}
	}

	return &projectMap, nil
}

// processDefaultBackend go through ingress default backend and updates the Open API Spec (openapi3.Swagger) of
// names.DefaultBackendProject in the ProjectsMap and the action Type to Update
func processDefaultBackend(ctx context.Context, reqInfo *common.RequestInfo, projects map[string]bool, projectMap *ProjectsMap, ing *v1beta1.Ingress) error {
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
			svc := &v1.Service{}
			if err := reqInfo.Client.Get(ctx, types.NamespacedName{Namespace: ing.Namespace, Name: ing.Spec.Backend.ServiceName}, svc); err != nil {
				if k8serrors.IsNotFound(err) {
					log.Error(err, "Service defined in the default backend is not found and skipping the default backend configuration in this ingress",
						"ingress", ing)
					// Skip this error without reconciling
					return nil
				}
				return err
			}

			u := urlFromIngBackend(ing, ing.Spec.Backend)
			pMap[names.DefaultBackendProject].Type = Update
			pMap[names.DefaultBackendProject].OAS.Servers = oasServers(u)
		}
	}

	return nil
}

// processIngressRules go through ingress rules and updates the Open API Spec (openapi3.Swagger) in
// the ProjectsMap and the action Type to Update
func processIngressRules(ctx context.Context, reqInfo *common.RequestInfo, projects map[string]bool, projectMap *ProjectsMap, ing *v1beta1.Ingress) error {
	log := reqInfo.Log
	pMap := *projectMap

	for _, rule := range ing.Spec.Rules {
		pj := names.HostToProject(rule.Host)

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

				svc := &v1.Service{}
				if err := reqInfo.Client.Get(ctx, types.NamespacedName{Namespace: ing.Namespace, Name: path.Backend.ServiceName}, svc); err != nil {
					if k8serrors.IsNotFound(err) {
						log.Error(err, "Service defined in the ingress rule path is not found and skipping the ingress rule path defined in the ingress",
							"ingress", ing, "rule_path", path)
						continue
					}
					return err
				}

				u := urlFromIngBackend(ing, &path.Backend)
				pMap[pj].OAS.Paths[oasPath] = oasPathItem(u)
			}
			// Update the project
			pMap[pj].Type = Update
		}
	}

	return nil
}

// processIngressTls go through ingress TLS rules and updates the Project.TlsCertificate in
// the ProjectsMap and the action Type to Update
func processIngressTls(ctx context.Context, reqInfo *common.RequestInfo, projects map[string]bool, projectMap *ProjectsMap, ing *v1beta1.Ingress) error {
	log := reqInfo.Log
	pMap := *projectMap

	for _, ingTls := range ing.Spec.TLS {
		// Check secret
		// TODO (renuka) do we want to verify the port
		secret := &v1.Secret{}
		if err := reqInfo.Client.Get(ctx, types.NamespacedName{Namespace: ing.Namespace, Name: ingTls.SecretName}, secret); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Error(err, "TLS secret not found and skipping TLS configuration for the hosts defined in the ingress", "ingress", ing, "hosts", ingTls.Hosts)
				// continue with other tls configs in the ingress with only skipping this config
				continue
			}
			return err
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
				pMap[pj].Type = Update
			}
		}
	}
	return nil
}

func tlsCertFromIngTls(secret *v1.Secret, ing *v1beta1.Ingress) (*TlsCertificate, error) {
	tlsCert := TlsCertificate{}

	crt, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("tls certificate not found in the field \"tls.crt\" of secret %s", secret.String()))
	}
	tlsCert.CertificateChain = crt

	key, ok := secret.Data["tls.key"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("tls key not found in the field \"tls.key\" of secret %s", secret.String()))
	}
	tlsCert.CertificateChain = key

	tlsConf := tls.Parse(ing)
	if tlsConf.TlsMode == tls.Mutual {
		caCert, ok := secret.Data["ca.crt"]
		if !ok {
			return nil, errors.New(fmt.Sprintf("tls ca cert not found in the field \"ca.crt\" of secret %s", secret.String()))
		}
		tlsCert.TrustedCa = caCert
	}

	return &tlsCert, nil
}

func urlFromIngBackend(ing *v1beta1.Ingress, backend *v1beta1.IngressBackend) string {
	tlsConf := tls.Parse(ing)
	protocol := "http"
	if tlsConf.TlsMode == tls.Origination || tlsConf.TlsMode == tls.Passthrough {
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
				"x-wso2-vhost": names.ProjectToHost(projectName),
				"x-wso2-spec":  "ingress",
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
