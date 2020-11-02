package configs

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"k8s.io/api/networking/v1beta1"
)

func Update(reqInfo *common.RequestInfo, ingresses []*v1beta1.Ingress) error {
	reqInfo.Log.Info("TEST")

	ingressConfigs, err := NewFromConfigMap(reqInfo)
	if err != nil {
		return err
	}

	ingressConfigs.Projects = []IngressConfig{
		{
			Ingress: "Test",
			Projects: []IngressProject{
				{
					Name: "foo_com",
					Host: "foo.com",
				},
			},
		},
	}

	_ = ingressConfigs.UpdateToConfigMap(reqInfo)

	// TODO (renuka)
	return nil
}
