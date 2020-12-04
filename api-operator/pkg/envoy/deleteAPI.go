package envoy

import (
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logDelete = log.Log.WithName("mgw.envoy.delete")

// TODO: Implement the delete API Function
func DeleteAPIFromMgw (client *client.Client, api *wso2v1alpha2.API) error {
	logDelete.Info("DELETED")
	return nil
}
