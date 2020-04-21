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
	mgwDockerFile          = "dockerfile-conf"
	interceptorsVolume     = "intcpt-vol"
	javaInterceptorsVolume = "java-intcpt-vol"

	analyticsCertFile       = "analytics-cert"
	analyticsCertLocation   = "/usr/wso2/analyticssecret/"
	analyticsVolumeName     = "analytics-volume-storage"
	analyticsVolumeLocation = "/home/ballerina/wso2/api-usage-data/"
	analyticsConfName       = "analytics-config"

	mgwToolkitImgConst   = "mgwToolkitImg"
	mgwRuntimeImgConst   = "mgwRuntimeImg"
	kanikoImgConst       = "kanikoImg"
	kanikoArguments      = "kanikoArguments"
	registryTypeConst    = "registryType"
	repositoryNameConst  = "repositoryName"
	wso2NameSpaceConst   = "wso2-system"
	policyConfigmap      = "policy-configmap"
	mgwConfSecretConst   = "mgw-conf"
	mgwConfConst         = "micro-gw.conf"
	controllerConfName   = "controller-config"
	ingressConfigs       = "ingress-configs"
	openShiftConfigs     = "route-configs"
	dockerRegConfigs     = "docker-registry-config"
	kanikoArgsConfigs    = "kaniko-arguments"
	ingressProperties    = "ingress.properties"
	routeProperties      = "route.properties"
	policyFileConst      = "policies.yaml"
	ingressResourceName  = "ingressResourceName"
	ingressTransportMode = "ingressTransportMode"
	ingressHostName      = "ingressHostName"
	tlsSecretName        = "tlsSecretName"
	routeName            = "routeName"
	routeHost            = "routeHost"
	routeTransportMode   = "routeTransportMode"
	tlsTermination       = "tlsTermination"

	defaultSecurity = "default-security-jwt"
	certPath        = "/usr/wso2/certs/"
	mgwConfMustache = "mgw-conf-mustache"
	mgwConfGoTmpl   = "mgwConf.gotmpl"

	portConst  = "port"
	httpConst  = "http"
	httpsConst = "https"

	operatorModeConst = "operatorMode"
	ingressMode       = "Ingress"
	clusterIPMode     = "ClusterIP"
	routeMode         = "Route"

	authorizationUrl       = "https://example.com/oauth/authorize"
	tokenUrl               = "https://example.com/oauth/token"
	oauthSecurityType      = "oauth2"
	basicSecurityType      = "http"
	basicSecurityAndScheme = "basic"

	securityJWT          = "JWT"
	securityOauth        = "Oauth"
	certAlias            = "alias"
	sidecar              = "sidecar"
	privateJet           = "privateJet"
	verifyHostNameVal    = "false"
	apiCrdDefaultVersion = "v1.0.0"

	hpaMaxReplicas                    = "hpaMaxReplicas"
	hpaTargetAverageUtilizationCPU    = "hpaTargetAverageUtilizationCPU"
	readinessProbeInitialDelaySeconds = "readinessProbeInitialDelaySeconds"
	readinessProbePeriodSeconds       = "readinessProbePeriodSeconds"
	livenessProbeInitialDelaySeconds  = "livenessProbeInitialDelaySeconds"
	livenessProbePeriodSeconds        = "livenessProbePeriodSeconds"

	resourceRequestCPU                 = "resourceRequestCPU"
	resourceRequestMemory              = "resourceRequestMemory"
	resourceLimitCPU                   = "resourceLimitCPU"
	resourceLimitMemory                = "resourceLimitMemory"
	generatekubernbetesartifactsformgw = "generatekubernbetesartifactsformgw"

	deploymentKind = "Deployment"
	serviceKind    = "Service"
	apiVersionKey  = "apps/v1"
	versionField   = "{version}"

	edge        = "edge"
	reencrypt   = "reencrypt"
	passthrough = "passthrough"
)
