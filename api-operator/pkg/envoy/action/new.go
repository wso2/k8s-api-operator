package action

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/names"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/api/networking/v1beta1"
	"strings"
)

func FromProjects(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress, projects map[string]bool) *ProjectsMap {
	log := reqInfo.Log
	projectMap := ProjectsMap{}
	// Initialize project action with type delete
	for p := range projects {
		projectMap[p] = &Project{
			Type: Delete,
			OAS:  defaultOpenAPI(p),
		}
	}

	for _, ing := range ingresses {
		if k8s.IsDeleted(ing) {
			// Ingress is deleted, no need to process
			continue
		}

		// check whether the ingress contributes to the default backend project
		if projects[names.DefaultBackendProject] && ing.Spec.Backend != nil {
			if projectMap[names.DefaultBackendProject].OAS.Servers != nil {
				// skip this default backend configuration
				// give priority to older ingress
				log.Info("Skipping the default backend configuration, since it is already defined by old ingress",
					"new_ingress", ing)
			} else {
				projectMap[names.DefaultBackendProject].Type = Update
				u := urlFromIngBackend(reqInfo.Namespace, ing.Spec.Backend)
				projectMap[names.DefaultBackendProject].OAS.Servers = oasServers(u)
			}
		}

		for _, rule := range ing.Spec.Rules {
			pj := names.HostToProject(rule.Host)

			// check whether the ingress contributes to the project
			if projects[pj] {
				projectMap[pj].Type = Update

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

					if projectMap[pj].OAS.Paths[oasPath] != nil {
						// skip this path
						// give priority to older ingress
						log.Info("Skipping the path configuration for the host defined, since it is already defined by old ingress",
							"new_ingress", ing, "host", rule.Host, "path", oasPath)
						continue
					}

					u := urlFromIngBackend(reqInfo.Namespace, &path.Backend)
					projectMap[pj].OAS.Paths[oasPath] = oasPathItem(u)
				}
			}
		}
	}

	return &projectMap
}

func urlFromIngBackend(namespace string, backend *v1beta1.IngressBackend) string {
	// TODO: (renuka) check TLS configs if not terminated should be HTTPS, use HTTP for now
	// Using only backend.ServiceName
	// TODO: (renuka) do validation for existence of service and throw error
	return fmt.Sprintf("http://%s.%s:%s", namespace, backend.ServiceName, backend.ServicePort.String())
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

func defaultOpenAPI(title string) *openapi3.Swagger {
	return &openapi3.Swagger{
		ExtensionProps: openapi3.ExtensionProps{},
		OpenAPI:        "3.0.0",
		Info: openapi3.Info{
			Title:       title,
			Description: title,
			Version:     "v1",
		},
		Servers:      nil,
		Paths:        openapi3.Paths{},
		Components:   openapi3.Components{},
		Security:     nil,
		ExternalDocs: nil,
	}
}
