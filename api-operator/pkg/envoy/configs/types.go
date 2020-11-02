package configs

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// TODO: (renuka) operatorNamespace represents the namespace of the operator
const operatorNamespace = "wso2-system"
const ingressConfigsCm = "ingress-configs"
const ingressConfigsKey = "ingress-configs"

// IngressConfigs represents a list of Open API Spec projects updated in the microgateway
type IngressConfigs struct {
	Projects []IngressConfig
}

// IngressConfig represents list of Open API Spec projects for an single ingress
type IngressConfig struct {
	Ingress  string
	Projects []IngressProject
}

// IngressProject represents an Open API Spec project name and related host
type IngressProject struct {
	Name string
	Host string
}

func (c *IngressConfigs) UpdateToConfigMap(reqInfo *common.RequestInfo) error {
	// Marshal yaml
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Check ingress-configs from configmap
	ingresCm := &v1.ConfigMap{}
	if err := (*reqInfo.Client).Get(reqInfo.Ctx, types.NamespacedName{
		Namespace: operatorNamespace, Name: ingressConfigsCm,
	}, ingresCm); err != nil {
		if !errors.IsNotFound(err) {
			// ConfigMap not found and create it.
			ingresCm.BinaryData = map[string][]byte{ingressConfigsKey: bytes}
			if err := (*reqInfo.Client).Create(reqInfo.Ctx, ingresCm); err != nil {
				return err
			}
		}
		return err
	}

	// Update ConfigMap
	d := ingresCm.BinaryData
	if d == nil {
		d = map[string][]byte{ingressConfigsKey: bytes}
	} else {
		d[ingressConfigsKey] = bytes
	}
	ingresCm.BinaryData = d
	if err := (*reqInfo.Client).Update(reqInfo.Ctx, ingresCm); err != nil {
		return err
	}
	return nil
}

// NewFromConfigMap returns a new IngressConfigs object with reading k8s config map
func NewFromConfigMap(reqInfo *common.RequestInfo) (*IngressConfigs, error) {
	// Fetch ingress-configs from configmap
	ingresCm := &v1.ConfigMap{}
	if err := (*reqInfo.Client).Get(reqInfo.Ctx, types.NamespacedName{
		Namespace: operatorNamespace, Name: ingressConfigsCm,
	}, ingresCm); err != nil {
		if !errors.IsNotFound(err) {
			return &IngressConfigs{}, nil
		}
		return nil, err
	}

	// Unmarshal to yaml
	configs := &IngressConfigs{}
	c := ingresCm.BinaryData[ingressConfigsKey]
	if err := yaml.Unmarshal(c, configs); err != nil {
		return nil, err
	}
	return configs, nil
}
