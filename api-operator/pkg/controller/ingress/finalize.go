package ingress

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
)

const (
	// finalizerName represents the name of ingress finalizer handled by this controller
	finalizerName = "wso2.microgateway/ingress.finalizer"
)

func (r *ReconcileIngress) finalizeDeletion(ctx context.Context, requestInfo *common.RequestInfo) error {
	// handle deletion with finalizers to avoid missing ingress configurations deleted while
	// restating controller, or deleted before starting controller.
	//
	// Ingress deletion delta change also handled in the update delta change flow and
	// skipping handling deletion here
	ingresses, err := getSortedIngressList(ctx, requestInfo)
	if err != nil {
		return err
	}

	if err := r.handleRequest(ctx, requestInfo, ingresses); err != nil {
		return nil
	}

	successfullyHandledRequestCount++
	return nil
}
