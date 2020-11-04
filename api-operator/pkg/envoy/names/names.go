package names

import (
	"fmt"
	"k8s.io/api/networking/v1beta1"
	"strings"
)

func HostToProject(host string) string {
	return fmt.Sprintf("ingress-%v", strings.ReplaceAll(host, ".", "_"))
}

func ProjectToHost(pj string) string {
	return strings.TrimPrefix(strings.ReplaceAll(pj, "_", "."), "ingress-")
}

func IngressToName(ing *v1beta1.Ingress) string {
	return fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
}
