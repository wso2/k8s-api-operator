package handler

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/client"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/status"
	"k8s.io/api/networking/v1beta1"
	"time"
)

type Handler struct {
	GatewayClient client.GatewayClient
}

func (h *Handler) UpdateWholeWorld(ctx context.Context, reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	reqInfo.Log.Info("Handle whole world update of the ingresses")

	// New state to be configured
	sDiff := status.NewFromIngresses(ingresses...)
	reqInfo.Log.V(1).Info("Changes in projects for ingresses", "new_status_changes", sDiff)

	return h.update(ctx, reqInfo, ingresses, sDiff)
}

func (h *Handler) UpdateDelta(ctx context.Context, reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	reqInfo.Log.Info("Handle delta update of the ingress")

	// New state to be configured
	instance := reqInfo.Object.(*v1beta1.Ingress)
	newS := status.NewFromIngresses(instance)
	reqInfo.Log.V(1).Info("Changes in projects for ingress", "new_status_changes", newS)

	return h.update(ctx, reqInfo, ingresses, newS)
}

func (h *Handler) update(ctx context.Context, reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress, sDiff *status.ProjectsStatus) error {
	// Read current state
	st, err := status.FromConfigMap(ctx, reqInfo)
	if err != nil {
		return err
	}
	reqInfo.Log.V(1).Info("Current status of Microgateway read from k8s configmap", "current_status", st)

	// Actions needed to happened with sDiff
	projectsSet := st.UpdatedProjects(sDiff)
	existingProjectSet := st.ProjectSet()
	reqInfo.Log.V(1).Info("Project set that require changes", "projects", projectsSet)
	projectsActions, err := action.FromProjects(ctx, reqInfo, ingresses, projectsSet, existingProjectSet)
	if err != nil {
		return err
	}

	reqInfo.Log.V(1).Info("Required actions on projects", "projects_actions", projectsActions)

	// Updated the gateway
	reqInfo.Log.Info("Updating projects on Microgateway")
	gatewayResponse, err := h.GatewayClient.Update(projectsActions)
	if err != nil {
		return err
	}
	reqInfo.Log.Info("Response from Microgateway on updating projects", "gateway_response", gatewayResponse)

	// Update the state back
	st.Update(sDiff, gatewayResponse)
	reqInfo.Log.V(1).Info("Updated state of gateway to be stored in k8s configmap based on the Microgateway response", "updated_current_state", st)

	// try update state without re handling request if error occurred
	var updateErr error
	for i := 0; i < 3; i++ {
		if updateErr = st.UpdateToConfigMap(ctx, reqInfo); updateErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if updateErr == nil {
		reqInfo.Log.V(1).Info("Successfully updated the updated_current_state in k8s configmap")
		reqInfo.Log.Info("Successfully updated the project")
	}
	return updateErr
}
