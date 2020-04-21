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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const (
	apimConfName       = "apim-config"
	wso2NameSpaceConst = "wso2-system"

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
	KeymanagerServerUrl string
	KeymanagerUsername  string
	KeymanagerPassword  string

	// jwtTokenConfig
	JwtConfigs *[]JwtTokenConfig

	// analytics
	AnalyticsEnabled          string
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
	KeymanagerServerUrl: "https://wso2apim.wso2:32001",
	KeymanagerUsername:  "admin",
	KeymanagerPassword:  "admin",

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
	AnalyticsEnabled:          "false",
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

// TODO: rnk: remove this after finish refactor
//var (
//
//	// Transport listener Configurations
//	httpPort           string = "9090"
//	httpsPort          string = "9095"
//	keystorePath       string = "${mgw-runtime.home}/runtime/bre/security/ballerinaKeystore.p12"
//	keystorePassword   string = "ballerina"
//	truststorePath     string = "${mgw-runtime.home}/runtime/bre/security/ballerinaTruststore.p12"
//	truststorePassword string = "ballerina"
//
//	//keymanager
//	KeymanagerServerurl string = "https://wso2apim.wso2:32001"
//	keymanagerUsername  string = "admin"
//	keymanagerPassword  string = "admin"
//
//	//jwtTokenConfig
//	Issuer           string = "https://wso2apim.wso2:32001/oauth2/token"
//	audience         string = "http://org.wso2.apimgt/gateway"
//	CertificateAlias string = "wso2apim310"
//
//	//analytics
//	analyticsEnabled          string = "false"
//	analyticsUsername         string = "admin"
//	analyticsPassword         string = "admin"
//	uploadingTimeSpanInMillis string = "600000"
//	rotatingPeriod            string = "600000"
//	uploadFiles               string = "true"
//	hostname                  string = "wso2apim.wso2"
//	port                      string = "32001"
//
//	//throttlingConfig
//	enabledGlobalTMEventPublishing string = "false"
//	jmsConnectionProvider          string = "wso2apim.wso2:5672"
//	throttleEndpoint               string = "wso2apim.wso2:32001"
//
//	//token revocation
//	enableRealtimeMessageRetrieval string = "false"
//
//	//validation
//	enableRequestValidation  string = "false"
//	enableResponseValidation string = "false"
//
//	//basic authentication
//	basicUsername string = "admin"
//	basicPassword string = "d033e22ae348aeb5660fc2140aec35850c4da997"
//
//	// HTTP client hostname verification
//	verifyHostname string = "true"
//
//	//log level
//	// TODO delete this, removed in micro-gateway 3.1.0
//	logLevel string = "INFO"
//)
