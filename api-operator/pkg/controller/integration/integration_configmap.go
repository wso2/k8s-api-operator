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
	"k8s.io/api/autoscaling/v2beta2"
	"sigs.k8s.io/yaml"
	"strconv"
)

// EIConfig control all deployments of the ei-ingress
type EIConfig struct {
	AutoCreateIngress bool
	SSLRedirect       string
	TLS               string
	Host              string
	EnableAutoScale   bool
	MinReplicas       int32
	MaxReplicas       int32
	HPAMetricSpec     []v2beta2.MetricSpec
}

// UpdateDefaultConfigs updates the default configs of Host, TLS, and ingress creation
func (r *ReconcileIntegration) UpdateDefaultConfigs(integration *wso2v1alpha1.Integration) EIConfig {
	eic := EIConfig {
		Host:              "wso2",
		AutoCreateIngress: true,
		SSLRedirect:       "true",
		EnableAutoScale:   false,
		MinReplicas:       1,
		MaxReplicas:       1,
		HPAMetricSpec:     nil,
	}

	var configMap, err = r.GetConfigMap(integration, nameForConfigMap())

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

		if configMap.Data[enableAutoScaleKey] != "" {
			eic.EnableAutoScale, err = strconv.ParseBool(configMap.Data[enableAutoScaleKey])
			if err != nil {
				log.Error(err, "Cannot parse enableAutoScaleKey to a boolean value. Setting false")
				eic.EnableAutoScale = false
			}
		}

		if configMap.Data[minReplicasKey] != "" {
			minReplicas, _ := strconv.ParseInt(configMap.Data[minReplicasKey],10,32)
			eic.MinReplicas = int32(minReplicas)
		}

		if configMap.Data[maxReplicasKey] != "" {
			maxReplicas, _ := strconv.ParseInt(configMap.Data[maxReplicasKey],10,32)
			eic.MaxReplicas = int32(maxReplicas)
		}

		if configMap.Data[hpaMetricsConfigKey] != "" {
			var hpaMetrics []v2beta2.MetricSpec
			yamlErr := yaml.Unmarshal([]byte(configMap.Data[hpaMetricsConfigKey]), &hpaMetrics)
			eic.HPAMetricSpec = hpaMetrics
			if yamlErr != nil {
				log.Error(yamlErr, "Error while reading HPAConfig")
			}
		}
	}
	return eic
}
