package names

import (
	"fmt"
	"strings"
)

func HostToProject(host string) string {
	return fmt.Sprintf("ingress-%v", strings.ReplaceAll(host, ".", "_"))
}

func ProjectToHost(pj string) string {
	return strings.TrimPrefix(strings.ReplaceAll(pj, "_", "."), "ingress-")
}
