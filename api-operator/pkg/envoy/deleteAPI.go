package envoy

import (
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteAPIFromMgw (client *client.Client, api *wso2v1alpha2.API) error {
	return nil
}
