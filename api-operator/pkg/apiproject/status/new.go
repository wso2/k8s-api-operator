// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package status

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject/names"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// FromConfigMap returns a new ProjectsStatus object with reading k8s config map
func FromConfigMap(ctx context.Context, reqInfo *common.RequestInfo) (*ProjectsStatus, error) {
	// Fetch ingress-status from configmap
	ingresCm := &v1.ConfigMap{}
	if err := reqInfo.Client.Get(ctx, types.NamespacedName{
		Namespace: config.SystemNamespace, Name: ingressProjectStatusCm,
	}, ingresCm); err != nil {
		if !errors.IsNotFound(err) {
			return &ProjectsStatus{}, nil
		}
		return &ProjectsStatus{}, nil
	}

	// Unmarshal to yaml
	st := &ProjectsStatus{}
	cm := ingresCm.Data[ingressProjectStatusKey]
	if err := yaml.Unmarshal([]byte(cm), st); err != nil {
		return nil, err
	}
	return st, nil
}

// NewFromIngresses returns a new ProjectsStatus from given Ingress objects
func NewFromIngresses(ingresses ...*ingress.Ingress) *ProjectsStatus {
	st := &ProjectsStatus{}
	for _, ing := range ingresses {
		updateFromIngress(st, ing)
	}
	return st
}

func updateFromIngress(projects *ProjectsStatus, ing *ingress.Ingress) {
	name := names.IngressToName(ing)
	(*projects)[name] = make(map[string]string)

	if k8s.IsDeleted(ing) {
		// Object being deleted
		// Should not affect the state
		return
	}

	if ing.Spec.Backend != nil {
		(*projects)[name][names.NoVHostProject] = "_"
	}

	// Projects for defined HTTP rules
	for _, rule := range ing.Spec.Rules {
		proj := names.HostToProject(rule.Host)
		(*projects)[name][proj] = "_"
	}

	// Projects for defined TLS rules
	for _, tls := range ing.Spec.TLS {
		for _, host := range tls.Hosts {
			proj := names.HostToProject(host)
			(*projects)[name][proj] = "_"
		}
	}
}
