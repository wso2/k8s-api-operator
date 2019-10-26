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
	dockerConfig       = "docker-config"
	mgwDockerFile      = "dockerfile-conf"
	swaggerVolume      = "swagger-volume"
	interceptorsVolume = "interceptors-volume"

	swaggerLocation            = "/usr/wso2/swagger/"
	dockerFileLocation         = "/usr/wso2/dockerfile/"
	dockerConfLocation         = "/kaniko/.docker"
	dockerFile                 = "dockerfile"
	dockerFileTemplate         = "dockerfile-template"
	policyyamlFile             = "policy-file"
	policyyamlLocation         = "/usr/wso2/policy/"
	mgwConfFile                = "conf-file"
	mgwConfLocation            = "/usr/wso2/mgwconf/"
	analyticsCertFile          = "analytics-cert"
	analyticsCertLocation      = "/usr/wso2/analyticssecret/"
	analyticsVolumeName        = "analytics-volume-storage"
	analyticsVolumeLocation    = "/home/ballerina/api-usage-data/"
	analyticsConfName          = "analytics-config"
	interceptorsVolumeLocation = "usr/wso2/interceptors/"

	mgwToolkitImgConst    = "mgwToolkitImg"
	mgwRuntimeImgConst    = "mgwRuntimeImg"
	kanikoImgConst        = "kanikoImg"
	dockerRegistryConst   = "dockerRegistry"
	wso2NameSpaceConst    = "wso2-system"
	policyConfigmap       = "policy-configmap"
	mgwConfSecretConst    = "mgw-conf"
	mgwConfConst          = "micro-gw.conf"
	dockerSecretNameConst = "docker-secret"
	controllerConfName    = "controller-config"
	policyFileConst       = "policies.yaml"

	usernameConst  = "username"
	passwordConst  = "password"
	certConst      = "cert_security"
	analyticsAlias = "wso2analytics260"

	dockerhubRegistryUrl                = "https://registry-1.docker.io/"
	defaultSecurity                     = "default-security-jwt"
	endpointExtension                   = "x-wso2-production-endpoints"
	deploymentMode                      = "x-wso2-mode"
	securityExtension                   = "security"
	certPath                            = "/usr/wso2/certs/"
	mgwConfTemplatePath                 = "/usr/local/bin/microgwconf.mustache"
	dockertemplatepath                  = "/usr/local/bin/dockerFile.gotmpl"
	mgwConfMustache                     = "mgw-conf-mustache"
	mgwConfGoTmpl                       = "mgwConf.gotmpl"
	dockerSecretMustache                = "docker-secret-mustache"
	dockerSecretMustacheTemplate        = "docketMustache.gotmpl"
	certConfig                          = "apim-certs"
	encodedTrustsorePassword            = "YmFsbGVyaW5h"
	truststoreSecretName                = "truststorepass"
	truststoreSecretData                = "password"
	apimConfName                        = "apim-config"
	keystorePathConst                   = "keystorePath"
	keystorePasswordConst               = "keystorePassword"
	truststorePathConst                 = "truststorePath"
	truststorePasswordConst             = "truststorePassword"
	keymanagerServerurlConst            = "keymanagerServerurl"
	keymanagerUsernameConst             = "keymanagerUsername"
	keymanagerPasswordConst             = "keymanagerPassword"
	issuerConst                         = "issuer"
	audienceConst                       = "audience"
	certificateAliasConst               = "certificateAlias"
	enabledGlobalTMEventPublishingConst = "enabledGlobalTMEventPublishing"
	basicUsernameConst                  = "basicUsername"
	basicPasswordConst                  = "basicPassword"
	analyticsEnabledConst               = "analyticsEnabled"
	analyticsUsernameConst              = "analyticsUsername"
	analyticsPasswordConst              = "analyticsPassword"
	uploadingTimeSpanInMillisConst      = "uploadingTimeSpanInMillis"
	rotatingPeriodConst                 = "rotatingPeriod"
	uploadFilesConst                    = "uploadFiles"
	verifyHostnameConst                 = "verifyHostname"
	hostnameConst                       = "hostname"
	portConst                           = "port"
	analyticsSecretConst                = "analyticsSecret"

	authorizationUrl        = "https://example.com/oauth/authorize"
	tokenUrl                = "https://example.com/oauth/token"
	oauthSecurityType       = "oauth2"
	basicSecurityType       = "http"
	basicSecurityAndScheme  = "basic"
	securitySchemeExtension = "securitySchemes"
	securityJWT             = "JWT"
	securityOauth           = "Oauth"
	certAlias               = "alias"
	pathsExtension          = "paths"
	sidecar                 = "sidecar"
	privateJet              = "privateJet"
	shared                  = "shared"
	verifyHostNameVal       = "false"

	hpaMaxReplicas                 = "hpaMaxReplicas"
	hpaTargetAverageUtilizationCPU = "hpaTargetAverageUtilizationCPU"
	readinessProbeInitialDelaySeconds = "readinessProbeInitialDelaySeconds"
	readinessProbePeriodSeconds = "readinessProbePeriodSeconds"
	livenessProbeInitialDelaySeconds = "livenessProbeInitialDelaySeconds"
	livenessProbePeriodSeconds = "livenessProbePeriodSeconds"

	resourceRequestCPU                 = "resourceRequestCPU"
	resourceRequestMemory              = "resourceRequestMemory"
	resourceLimitCPU                   = "resourceLimitCPU"
	resourceLimitMemory                = "resourceLimitMemory"
	generatekubernbetesartifactsformgw = "generatekubernbetesartifactsformgw"
)
