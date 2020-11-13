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
package targetendpoint

const (
	privateJet = "privateJet"
	serverless = "serverless"

	hpaConfigMapName        = "hpa-configs"
	maxReplicasConfigKey    = "targetEndpointMaxReplicas"
	metricsConfigKey        = "targetEndpointMetrics"
	metricsConfigKeyV2beta1 = "targetEndpointMetricsV2beta1"
	hpaVersionConst         = "hpaVersion"

	resourceRequestCPUTarget    = "resourceRequestCPUTarget"
	resourceRequestMemoryTarget = "resourceRequestMemoryTarget"
	resourceLimitCPUTarget      = "resourceLimitCPUTarget"
	resourceLimitMemoryTarget   = "resourceLimitMemoryTarget"

	portKey = "port"

	deploymentKind    = "Deployment"
	serviceKind       = "Service"
	apiVersion        = "apps/v1"
	knativeApiVersion = "serving.knative.dev/v1"
)
