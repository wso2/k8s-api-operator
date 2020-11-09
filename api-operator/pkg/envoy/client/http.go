package client

import "github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"

type Http struct {
}

func (c *Http) Update(projects *action.ProjectsMap) (Response, error) {
	// TODO (renuka) call HTTP client

	// sample response
	//r := Response{
	//	"ingress-__bar_org":    Updated,
	//	"ingress-__foo_org":    Failed,
	//	"ingress-prod_foo_org": Updated,
	//}

	return NewFakeAllSucceeded().Update(projects)
}
