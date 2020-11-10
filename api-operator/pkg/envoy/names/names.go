package names

import (
	"fmt"
	"k8s.io/api/networking/v1beta1"
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

func IngressToName(ing *v1beta1.Ingress) string {
	return fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
}
