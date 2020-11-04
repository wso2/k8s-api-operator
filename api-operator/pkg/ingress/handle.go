package ingress

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/controller"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/status"
	"k8s.io/api/networking/v1beta1"
)

func UpdateDelta(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	log := reqInfo.Log
	log.Info("Handle delta update of the ingress")

	// Read current state
	st, err := status.FromConfigMap(reqInfo)
	if err != nil {
		return err
	}

	// New state to be configured
	instance := reqInfo.Object.(*v1beta1.Ingress)
	newSt := status.NewFromIngress(instance)
	projectsList := st.UpdatedProjects(newSt)

	projectsActions := action.FromProjects(reqInfo, ingresses, projectsList)

	_, err = controller.UpdateGateway(projectsActions) // gatewayResponse
	if err != nil {
		return err
	}

	ingressConfigs := &status.ProjectsStatus{"ing1": map[string]string{"foo_com": "_", "bar_com": "_"}}

	_ = ingressConfigs.UpdateToConfigMap(reqInfo)

	// TODO (renuka)
	return nil
}
