/*
 * Copyright (c) 2020 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package integration

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
)

// serviceForIntegration returns a service object
func (r *ReconcileIntegration) serviceForIntegration(m *wso2v1alpha1.Integration) *corev1.Service {
	//set HTTP and HTTPS ports for as ServiceSpec ports
	exposeServicePorts := []corev1.ServicePort{
		corev1.ServicePort{
			Name:       m.Name + strconv.Itoa(8290),
			Port:       8290,
			TargetPort: intstr.FromInt(8290),
		},
	}

	// check inbound endpoint port is exist and append to the container port
	for _, port := range m.Spec.InboundPorts {
		exposeServicePorts = append(
			exposeServicePorts,
			corev1.ServicePort{
				Name: m.Name + strconv.Itoa(int(port)),
				Port: port,
				TargetPort: intstr.IntOrString{
					Type:   Int,
					IntVal: port,
				},
			},
		)
	}

	labels := labelsForIntegration(m.Name)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameForService(m),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    exposeServicePorts,
		},
	}
	// Set Integration instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}
