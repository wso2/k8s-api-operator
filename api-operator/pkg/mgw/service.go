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

package mgw

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	ingressMode   = "Ingress"
	routeMode     = "Route"
	clusterIPMode = "ClusterIP"

	portConst  = "port"
	httpConst  = "http"
	httpsConst = "https"
)

//Creating a LB balancer service to expose mgw
func Service(api *wso2v1alpha1.API, operatorMode string, owner []metav1.OwnerReference) *corev1.Service {
	var serviceType corev1.ServiceType
	serviceType = corev1.ServiceTypeLoadBalancer

	if strings.EqualFold(operatorMode, ingressMode) || strings.EqualFold(operatorMode, clusterIPMode) ||
		strings.EqualFold(operatorMode, routeMode) {
		serviceType = corev1.ServiceTypeClusterIP
	}

	labels := map[string]string{
		"app": api.Name,
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            api.Name,
			Namespace:       api.Namespace,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{{
				Name:       httpsConst + "-" + portConst,
				Port:       Configs.HttpsPort,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpsPort},
			}, {
				Name:       httpConst + "-" + portConst,
				Port:       Configs.HttpPort,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpPort},
			}},
			Selector: labels,
		},
	}

	return svc
}
