package handler

import (
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

func (h *Handler) UpdateWholeWorld(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	log := reqInfo.Log
	log.Info("Handle whole world update of the ingresses")

	// New state to be configured
	newS := status.NewFromIngresses(ingresses...)

	return h.update(reqInfo, ingresses, newS)
}

func (h *Handler) UpdateDelta(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	log := reqInfo.Log
	log.Info("Handle delta update of the ingress")

	// New state to be configured
	instance := reqInfo.Object.(*v1beta1.Ingress)
	newS := status.NewFromIngresses(instance)

	return h.update(reqInfo, ingresses, newS)
}

func (h *Handler) update(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress, newStatus *status.ProjectsStatus) error {
	// Read current state
	st, err := status.FromConfigMap(reqInfo)
	if err != nil {
		return err
	}

	// Actions needed to happened with newStatus
	projectsSet := st.UpdatedProjects(newStatus)
	projectsActions := action.FromProjects(reqInfo, ingresses, projectsSet)

	// Updated the gateway
	gatewayResponse, err := h.GatewayClient.Update(projectsActions)
	if err != nil {
		return err
	}

	// Update the state back
	st.Update(newStatus, gatewayResponse)

	// try update state without re handling request if error occurred
	var updateErr error
	for i := 0; i < 3; i++ {
		if updateErr = st.UpdateToConfigMap(reqInfo); updateErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return updateErr
}
