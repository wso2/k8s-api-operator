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
	"strconv"
)

// EIConfigNew bares all configurations related to EI deployment
type EIConfigNew struct {
	integration          wso2v1alpha1.Integration
	integrationConfigMap corev1.ConfigMap
	ingressConfigMap     corev1.ConfigMap
}

// PopulateConfigurations updates the default configs of Host, TLS, and ingress creation. Read from Integration first,
// If not defined, read defaults from integrationConfigMap and update integration
func (r *ReconcileIntegration) PopulateConfigurations(integration *wso2v1alpha1.Integration) (EIConfigNew, error) {

	var integrationConfigMap, err = r.GetConfigMap(integration, integrationConfigMapName)
	if err != nil {
		log.Error(err, "There is no integration config-map found, or there's an error reading it")
		return EIConfigNew{}, err
	}

	var ingressConfigMap, err1 = r.GetConfigMap(integration, integrationIngressConfigMapName)
	if err1 != nil {
		log.Error(err, "There is no integration config-map found, or there's an error reading it")
		return EIConfigNew{}, err1
	}

	if integration.Spec.DeploySpec.MinReplicas == 0 {
		if integrationConfigMap.Data[minReplicasKey] != "" {
			minReplicas, err := strconv.ParseInt(integrationConfigMap.Data[minReplicasKey], 10, 32)
			if err != nil {
				log.Error(err, "Cannot parse minReplicasKey to a int value.")
				return EIConfigNew{}, err
			}
			integration.Spec.DeploySpec.MinReplicas = int32(minReplicas)
		}
	}

	if integration.Spec.DeploySpec.ReqCpu == "" {
		if integrationConfigMap.Data[requestCPUKey] != "" {
			integration.Spec.DeploySpec.ReqCpu = integrationConfigMap.Data[requestCPUKey]
		}
	}

	if integration.Spec.DeploySpec.ReqMemory == "" {
		if integrationConfigMap.Data[reqMemoryKey] != "" {
			integration.Spec.DeploySpec.ReqMemory = integrationConfigMap.Data[reqMemoryKey]
		}
	}

	if integration.Spec.DeploySpec.LimitCpu == "" {
		if integrationConfigMap.Data[cpuLimitKey] != "" {
			integration.Spec.DeploySpec.LimitCpu = integrationConfigMap.Data[cpuLimitKey]
		}
	}

	if integration.Spec.DeploySpec.MemoryLimit == "" {
		if integrationConfigMap.Data[memoryLimitKey] != "" {
			integration.Spec.DeploySpec.MemoryLimit = integrationConfigMap.Data[memoryLimitKey]
		}
	}

	if integration.Spec.AutoScale.Enabled == "" {
		autoScaleEnabled, err := strconv.ParseBool(integrationConfigMap.Data[enableAutoScaleKey])
		if err != nil {
			log.Error(err, "Cannot parse enableAutoScaleKey to a boolean value. Setting false")
			integration.Spec.AutoScale.Enabled = strconv.FormatBool(false)
		} else {
			integration.Spec.AutoScale.Enabled = strconv.FormatBool(autoScaleEnabled)
		}
	}

	if integration.Spec.AutoScale.MaxReplicas == 0 {
		if integrationConfigMap.Data[maxReplicasKey] != "" {
			maxReplicas, err := strconv.ParseInt(integrationConfigMap.Data[maxReplicasKey], 10, 32)
			if err != nil {
				log.Error(err, "Cannot parse minReplicasKey to a int value.")
				return EIConfigNew{}, err
			}
			integration.Spec.AutoScale.MaxReplicas = int32(maxReplicas)
		}
	}

	if integration.Spec.Expose.PassthroPort == 0 {
		integration.Spec.Expose.PassthroPort = defaultPassthroPort
	}

	eiConfig := EIConfigNew {
		integration:          *integration,
		integrationConfigMap: *integrationConfigMap,
		ingressConfigMap: *ingressConfigMap,
	}

	return eiConfig, nil
}
