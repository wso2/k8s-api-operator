package ingress

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/status"
	"k8s.io/api/networking/v1beta1"
)

func Update(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	reqInfo.Log.Info("TEST")
	//instance := reqInfo.Object.(*v1beta1.Ingress)

	ingressConfigs, err := status.NewFromConfigMap(reqInfo)
	if err != nil {
		return err
	}

	ingressConfigs = &status.ProjectsStatus{"ing1": map[string]string{"foo_com": "_", "bar_com": "_"}}

	_ = ingressConfigs.UpdateToConfigMap(reqInfo)

	// TODO (renuka)
	return nil
}
