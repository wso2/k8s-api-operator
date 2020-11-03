package status

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/names"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// FromConfigMap returns a new ProjectsStatus object with reading k8s config map
func FromConfigMap(reqInfo *common.RequestInfo) (*ProjectsStatus, error) {
	// Fetch ingress-status from configmap
	ingresCm := &v1.ConfigMap{}
	if err := (*reqInfo.Client).Get(reqInfo.Ctx, types.NamespacedName{
		Namespace: operatorNamespace, Name: ingressProjectStatusCm,
	}, ingresCm); err != nil {
		if !errors.IsNotFound(err) {
			return &ProjectsStatus{}, nil
		}
		return &ProjectsStatus{}, nil
	}

	// Unmarshal to yaml
	st := &ProjectsStatus{}
	cm := ingresCm.BinaryData[ingressProjectStatusKey]
	if err := yaml.Unmarshal(cm, st); err != nil {
		return nil, err
	}
	return st, nil
}

// NewFromIngress returns a new ProjectsStatus from given Ingress object
func NewFromIngress(ing *v1beta1.Ingress) *ProjectsStatus {
	st := &ProjectsStatus{}
	(*st)[ing.Name] = make(map[string]string)

	// Projects for defined HTTP rules
	for _, rule := range ing.Spec.Rules {
		proj := names.HostToProject(rule.Host)
		(*st)[ing.Name][proj] = "_"
	}

	// Projects for defined TLS rules
	// TODO: (renuka) handle TLS

	return st
}
