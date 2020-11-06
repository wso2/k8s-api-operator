package names

import (
	"fmt"
	"k8s.io/api/networking/v1beta1"
	"strings"
)

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
