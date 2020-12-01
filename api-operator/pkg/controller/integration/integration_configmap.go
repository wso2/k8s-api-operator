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
	"context"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EIController control all deployments of the ei-ingress
type EIController struct {
	AutoCreateIngress bool
	SSLRedirect       string
	TLS               string
	Host              string
}

// UpdateDefaultConfigs updates the default configs of Host, TLS, and ingress creation
func (r *ReconcileIntegration) UpdateDefaultConfigs(integration *wso2v1alpha1.Integration) EIController {
	eic := EIController{
		Host:              "wso2",
		AutoCreateIngress: true,
		SSLRedirect:       "true",
	}

	configMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: nameForConfigMap(), Namespace: integration.Namespace}, configMap)

	if err == nil {
		if configMap.Data["host"] != "" {
			eic.Host = configMap.Data["host"]
		}

		if configMap.Data["autoIngressCreation"] != "" {
			if configMap.Data["autoIngressCreation"] == "true" {
				eic.AutoCreateIngress = true
			} else {
				eic.AutoCreateIngress = false
			}
		}

		if configMap.Data["sslRedirect"] != "" {
			if configMap.Data["sslRedirect"] == "true" {
				eic.SSLRedirect = "true"
			} else {
				eic.SSLRedirect = "false"
			}
		}

		if configMap.Data["ingressTLS"] != "" {
			eic.TLS = configMap.Data["ingressTLS"]
		}
	}
	return eic
}
