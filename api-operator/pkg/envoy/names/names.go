package names

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"strings"
)

// DefaultBackendProject represents the project of default backend of ingress
// project for a host in an ingress rule can not have three "_" consecutively.
// So this name is not conflict with a project for a host in an ingress rule.
const DefaultBackendProject = "ingress-___default"

func HostToProject(host string) string {
	p := strings.ReplaceAll(host, "*.", "__")
	return fmt.Sprintf("ingress-%v", strings.ReplaceAll(p, ".", "_"))
}

func ProjectToHost(pj string) string {
	p := strings.TrimPrefix(strings.ReplaceAll(pj, "__", "*."), "ingress-")
	return strings.ReplaceAll(p, "_", ".")
}

func IngressToName(ing *ingress.Ingress) string {
	return fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
}
