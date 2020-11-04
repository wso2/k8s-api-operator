package controller

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
)

func UpdateGateway(projects *action.ProjectsMap) (Response, error) {
	// TODO (renuka) call HTTP client

	// sample response
	r := Response{
		"ingress-foo_com": 200,
		"ingress-bar_com": 500,
	}

	return r, nil
}
