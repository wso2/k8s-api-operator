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
package api

const (
	mgwToolkitImgConst  = "mgwToolkitImg"
	mgwRuntimeImgConst  = "mgwRuntimeImg"
	kanikoArguments     = "kanikoArguments"
	registryTypeConst   = "registryType"
	repositoryNameConst = "repositoryName"
	controllerConfName  = "controller-config"
	ingressConfigs      = "ingress-configs"
	openShiftConfigs    = "route-configs"
	dockerRegConfigs    = "docker-registry-config"
	kanikoArgsConfigs   = "kaniko-arguments"

	dockerPushRegName        = "pushRegistryName"
	imagePullSecretNameConst = "imagePullSecretName"

	operatorModeConst             = "operatorMode"
	istioMode                     = "Istio"
	ingressMode                   = "Ingress"
	routeMode                     = "Route"
	observabilityEnabledConfigKey = "observabilityEnabled"

	eventTypeError = "Error"

	ingressHostName = "ingressHostName"
	routeHost       = "routeHost"

	sidecar                           = "sidecar"
	apiCrdDefaultVersion              = "v1.0.0"
	generateKubernetesArtifactsForMgw = "generatekubernbetesartifactsformgw"

	hpaConfigMapName = "hpa-configs"
	hpaVersionConst  = "hpaVersion"
)
