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
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const (
	apimConfName       = "apim-config"
	mgwConfMustache    = "mgw-conf-mustache"
	wso2NameSpaceConst = "wso2-system"

	mgwConfGoTmpl      = "mgwConf.gotmpl"
	mgwConfSecretConst = "mgw-conf"
	mgwConfConst       = "micro-gw.conf"
	mgwConfLocation    = "/usr/wso2/mgwconf/"

	httpPortValConst  = 9090
	httpsPortValConst = 9095
)

const (
	verifyHostnameConst                 = "verifyHostname"
	enabledGlobalTMEventPublishingConst = "enabledGlobalTMEventPublishing"
	jmsConnectionProviderConst          = "jmsConnectionProvider"
	throttleEndpointConst               = "throttleEndpoint"
	enableRealtimeMessageRetrievalConst = "enableRealtimeMessageRetrieval"
	enableRequestValidationConst        = "enableRequestValidation"
	enableResponseValidationConst       = "enableResponseValidation"
	logLevelConst                       = "logLevel"
	httpPortConst                       = "httpPort"
	httpsPortConst                      = "httpsPort"
)

type Configuration struct {
	// transport listener Configurations
	HttpPort           int32
	HttpsPort          int32
	KeystorePath       string
	KeystorePassword   string
	TruststorePath     string
	TruststorePassword string

	// key manager
	KeyManagerServerUrl string
	KeyManagerUsername  string
	KeyManagerPassword  string

	// jwtTokenConfig
	JwtConfigs *[]JwtTokenConfig

	// analytics
	AnalyticsEnabled          bool
	AnalyticsUsername         string
	AnalyticsPassword         string
	UploadingTimeSpanInMillis string
	RotatingPeriod            string
	UploadFiles               string
	AnalyticsHostname         string
	AnalyticsPort             string

	// throttlingConfig
	EnabledGlobalTMEventPublishing string
	JmsConnectionProvider          string
	ThrottleEndpoint               string

	// token revocation
	EnableRealtimeMessageRetrieval string

	// validation
	EnableRequestValidation  string
	EnableResponseValidation string

	//basic authentication
	BasicUsername string
	BasicPassword string

	// HTTP client hostname verification
	VerifyHostname string

	//log level
	LogLevel string
}

type JwtTokenConfig struct {
	CertificateAlias     string
	Issuer               string
	Audience             string
	ValidateSubscription bool
}

// mgw configs with default values
var Configs = &Configuration{
	// transport listener Configurations
	HttpPort:           9090,
	HttpsPort:          9095,
	KeystorePath:       "${mgw-runtime.home}/runtime/bre/security/ballerinaKeystore.p12",
	KeystorePassword:   "ballerina",
	TruststorePath:     "${mgw-runtime.home}/runtime/bre/security/ballerinaTruststore.p12",
	TruststorePassword: "ballerina",

	// key manager
	KeyManagerServerUrl: "https://wso2apim.wso2:32001",
	KeyManagerUsername:  "admin",
	KeyManagerPassword:  "admin",

	// jwtTokenConfig
	JwtConfigs: &[]JwtTokenConfig{
		{
			CertificateAlias:     "wso2apim310",
			Issuer:               "https://wso2apim.wso2:32001/oauth2/token",
			Audience:             "http://org.wso2.apimgt/gateway",
			ValidateSubscription: false,
		},
	},

	// analytics
	AnalyticsEnabled:          false,
	AnalyticsUsername:         "admin",
	AnalyticsPassword:         "admin",
	UploadingTimeSpanInMillis: "600000",
	RotatingPeriod:            "600000",
	UploadFiles:               "true",
	AnalyticsHostname:         "wso2apim.wso2",
	AnalyticsPort:             "32001",

	// throttlingConfig
	EnabledGlobalTMEventPublishing: "false",
	JmsConnectionProvider:          "wso2apim.wso2:5672",
	ThrottleEndpoint:               "wso2apim.wso2:32001",

	// token revocation
	EnableRealtimeMessageRetrieval: "false",

	// validation
	EnableRequestValidation:  "false",
	EnableResponseValidation: "false",

	//basic authentication
	BasicUsername: "admin",
	BasicPassword: "d033e22ae348aeb5660fc2140aec35850c4da997",

	// HTTP client hostname verification
	VerifyHostname: "true",

	//log level
	LogLevel: "INFO",
}

// SetApimConfigs sets the MGW configs from APIM configmap
func SetApimConfigs(client *client.Client) error {
	// get data from APIM configmap
	apimConfig := k8s.NewConfMap()
	errApim := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: apimConfName}, apimConfig)

	if errApim != nil {
		if errors.IsNotFound(errApim) {
			logger.Info("APIM config is not found. Continue with default configs")
		} else {
			logger.Error(errApim, "Error retrieving APIM configs")
			return errApim
		}
	}

	Configs.VerifyHostname = apimConfig.Data[verifyHostnameConst]
	Configs.EnabledGlobalTMEventPublishing = apimConfig.Data[enabledGlobalTMEventPublishingConst]
	Configs.JmsConnectionProvider = apimConfig.Data[jmsConnectionProviderConst]
	Configs.ThrottleEndpoint = apimConfig.Data[throttleEndpointConst]
	Configs.EnableRealtimeMessageRetrieval = apimConfig.Data[enableRealtimeMessageRetrievalConst]
	Configs.EnableRequestValidation = apimConfig.Data[enableRequestValidationConst]
	Configs.EnableResponseValidation = apimConfig.Data[enableResponseValidationConst]
	Configs.LogLevel = apimConfig.Data[logLevelConst]
	httpPort, err := strconv.Atoi(apimConfig.Data[httpPortConst])
	if err != nil {
		logger.Error(err, "Provided http port is not valid. Using the default port")
		Configs.HttpPort = httpPortValConst
	} else {
		Configs.HttpPort = int32(httpPort)
	}
	httpsPort, err := strconv.Atoi(apimConfig.Data[httpsPortConst])
	if err != nil {
		logger.Error(err, "Provided https port is not valid. Using the default port")
		Configs.HttpsPort = httpsPortValConst
	} else {
		Configs.HttpsPort = int32(httpsPort)
	}

	return nil
}

// ApplyConfFile render and add the MGW configuration file to cluster
func ApplyConfFile(client *client.Client, userNamespace, apiName string, owner *[]metav1.OwnerReference) error {
	// retrieving the MGW template configmap
	templateConfMap := k8s.NewConfMap()
	errConf := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: mgwConfMustache}, templateConfMap)
	if errConf != nil {
		logger.Error(errConf, "Error retrieving the MGW template configmap")
		return errConf
	}

	// render final micro-gw-conf file
	templateText := templateConfMap.Data[mgwConfGoTmpl]
	finalConf, errRender := str.RenderTemplate(templateText, Configs)
	if errRender != nil {
		logger.Error(errRender, "Error rendering the MGW configuration file")
		return errRender
	}

	// create MGW config file as a secret in the k8s cluster
	confNsName := types.NamespacedName{Namespace: userNamespace, Name: apiName + "-" + mgwConfSecretConst}
	confData := map[string][]byte{mgwConfConst: []byte(finalConf)}
	confSecret := k8s.NewSecretWith(confNsName, &confData, nil, owner)
	if err := k8s.Apply(client, confSecret); err != nil {
		return err
	}

	// add volumes to Kaniko job
	kaniko.AddVolume(k8s.SecretVolumeMount(confNsName.Name, mgwConfLocation))
	return nil
}
