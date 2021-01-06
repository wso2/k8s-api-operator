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

const (

	deploymentAPIVersion = "apps/v1"
	deploymentKind = "Deployment"

	integrationControllerName       = "integration-controller"
	integrationConfigMapName        = "integration-config"
	integrationIngressConfigMapName = "integration-ingress-config"
	eiContainerName                 = "micro-integrator"
	defaultPassthroPort             = 8290

	deploymentNamePostfix = "-deployment"
	hpaNamePostfix = "-hpa"
	serviceNamePostfix = "-service"
	inboundServicePostfix = "-inbound"
	eiIngressName = "ei-operator-ingress"

	ingressHostNameKey = "ingressHostName"
	autoIngressCreationKey = "autoIngressCreation"
	ingressProperties    = "ingress.properties"
	sslRedirectKey = "sslRedirect"
	tlsSecretNameKey = "tlsSecretName"


	requestCPUKey = "requestCPU"
	reqMemoryKey = "reqMemory"
	cpuLimitKey = "cpuLimit"
	memoryLimitKey = "memoryLimit"

	hpaMetricsConfigKey = "hpaMetrics"
	enableAutoScaleKey = "enableAutoScale"
	minReplicasKey = "minReplicas"
	maxReplicasKey = "maxReplicas"
)
