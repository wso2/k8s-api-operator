package ingress

import "github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"

const (
	// finalizerName represents the name of ingress finalizer handled by this controller
	finalizerName = "wso2.microgateway/ingress.finalizer"
)

func finalizeDeletion(requestInfo *common.RequestInfo) error {
	// TODO (renuka)
	return nil
}
