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
	"strconv"
	"strings"

	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logConf = log.Log.WithName("mgw.config")

const (
	apimConfName    = "apim-config"
	apimSecretName  = "apim-secret"
	mgwConfMustache = "mgw-conf-mustache"

	mgwConfGoTmpl      = "mgwConf.gotmpl"
	mgwConfSecretConst = "mgw-conf"
	mgwConfConst       = "micro-gw.conf"
	mgwConfLocation    = "/usr/wso2/mgwconf/"

	httpPortValConst  = 9090
	httpsPortValConst = 9095

	validityTimeValConst = -1

	observabilityPrometheusPort = 9797

	jwtTokenExpiryValConst              = 900000
	jwtTokenCacheExpiryTimeValConst     = 900000
	jwtTokenCacheCapacityValConst       = 10000
	jwtTokenCacheEvictionFactorValConst = 0.25
)

const (
	verifyHostnameConst                 = "verifyHostname"
	enabledGlobalTMEventPublishingConst = "enabledGlobalTMEventPublishing"
	jmsConnectionProviderConst          = "jmsConnectionProvider"
	throttleEndpointConst               = "throttleEndpoint"
	throttlingRecieverURLConst          = "throttlingRecieverUrl"
	throttlingAuthURLConst              = "throttlingAuthUrl"
	enableRealtimeMessageRetrievalConst = "enableRealtimeMessageRetrieval"
	enableRequestValidationConst        = "enableRequestValidation"
	enableResponseValidationConst       = "enableResponseValidation"
	enabledEventhub                     = "enabledEventhub"
	logLevelConst                       = "logLevel"
	httpPortConst                       = "httpPort"
	httpsPortConst                      = "httpsPort"
	enabledAPIKeyIssuerConst            = "enabledAPIKeyIssuer"
	apiKeyKeystorePathConst             = "apiKeyKeystorePath"
	apiKeyKeystorePasswordConst         = "apiKeyKeystorePassword"
	apiKeyIssuerNameConst               = "apiKeyIssuerName"
	apiKeyIssuerCertificateAliasConst   = "apiKeyIssuerCertificateAlias"
	validityTimeConst                   = "validityTime"
	allowedAPIsConst                    = "allowedAPIs"
	jwtHeaderConst                      = "jwtHeader"
	enabledJwtGeneratorConst            = "enabledJwtGenerator"
	jwtClaimDialectConst                = "jwtClaimDialect"
	jwtSigningAlgorithmConst            = "jwtSigningAlgorithm"
	jwtCertificateAliasConst            = "jwtCertificateAlias"
	jwtPrivateKeyAliasConst             = "jwtPrivateKeyAlias"
	jwtTokenExpiryConst                 = "jwtTokenExpiry"
	jwtRestrictedClaimsConst            = "jwtRestrictedClaims"
	jwtIssuerConst                      = "jwtIssuer"
	jwtAudienceConst                    = "jwtAudience"
	jwtGeneratorImplConst               = "jwtGeneratorImpl"
	jwtTokenCacheEnableConst            = "jwtTokenCacheEnable"
	jwtTokenCacheExpiryTimeConst        = "jwtTokenCacheExpiryTime"
	jwtTokenCacheCapacityConst          = "jwtTokenCacheCapacity"
	jwtTokenCacheEvictionFactorConst    = "jwtTokenCacheEvictionFactor"
)

// Configuration represents configurations for MGW
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
	ApimUsername                   string
	ApimPassword                   string
	ThrottlingRecieverURL          string
	ThrottlingAuthURL              string

	// token revocation
	EnableRealtimeMessageRetrieval string

	// validation
	EnableRequestValidation  string
	EnableResponseValidation string
	EnabledEventhub          string

	//basic authentication
	BasicUsername string
	BasicPassword string

	// HTTP client hostname verification
	VerifyHostname string

	// log level
	LogLevel string

	//APIKeyIssuerConfig
	EnabledAPIKeyIssuer          string
	APIKeyKeystorePath           string
	APIKeyKeystorePassword       string
	APIKeyIssuerName             string
	APIKeyIssuerCertificateAlias string
	ValidityTime                 int32

	// APIKeyTokenConfig
	APIKeyConfigs *[]APIKeyTokenConfig

	//APIKey Allowed API names and versions
	APIKeyAllowedAPIs APIKeyTokenAllowedAPIs

	// enable observability of MGW
	ObservabilityEnabled bool

	// JWT Header Configs
	JwtHeader string

	// JWT Generator Configs
	EnabledJwtGenerator         string
	JwtClaimDialect             string
	JwtSigningAlgorithm         string
	JwtCertificateAlias         string
	JwtPrivateKeyAlias          string
	JwtTokenExpiry              int32
	JwtRestrictedClaims         []string
	JwtIssuer                   string
	JwtAudience                 []string
	JwtGeneratorImpl            string
	JwtTokenCacheEnable         string
	JwtTokenCacheExpiryTime     int32
	JwtTokenCacheCapacity       int32
	JwtTokenCacheEvictionFactor float64
}

type JwtTokenConfig struct {
	CertificateAlias     string
	Issuer               string
	Audience             string
	ValidateSubscription bool
	AudiencePresent      bool
	JwksPresent          bool
	JwksURL              string
}

type APIKeyTokenConfig struct {
	APIKeyCertificateAlias string
	APIKeyIssuer           string
	APIKeyAudience         string
	ValidateAllowedAPIs    bool
}

type APIKeyTokenAllowedAPIs []map[string]string

// Configs represents MGW configs with default values
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
			CertificateAlias:     "wso2apim320",
			Issuer:               "https://wso2apim.wso2:32001/oauth2/token",
			Audience:             "http://org.wso2.apimgt/gateway",
			ValidateSubscription: false,
			AudiencePresent:      false,
			JwksPresent:          false,
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
	ApimUsername:                   "admin",
	ApimPassword:                   "admin",
	ThrottlingRecieverURL:          "wso2apim.wso2:9611",
	ThrottlingAuthURL:              "wso2apim.wso2:9611",

	// token revocation
	EnableRealtimeMessageRetrieval: "false",

	// validation
	EnableRequestValidation:  "false",
	EnableResponseValidation: "false",
	EnabledEventhub:          "false",

	//basic authentication
	BasicUsername: "admin",
	BasicPassword: "d033e22ae348aeb5660fc2140aec35850c4da997",

	// HTTP client hostname verification
	VerifyHostname: "true",

	// log level
	LogLevel: "INFO",

	//APIKeyIssuerConfig
	EnabledAPIKeyIssuer:          "true",
	APIKeyKeystorePath:           "${mgw-runtime.home}/runtime/bre/security/ballerinaKeystore.p12",
	APIKeyKeystorePassword:       "ballerina",
	APIKeyIssuerName:             "https://localhost:9095/apikey",
	APIKeyIssuerCertificateAlias: "ballerina",
	ValidityTime:                 -1,

	// APIKeyTokenConfig
	APIKeyConfigs: &[]APIKeyTokenConfig{
		{
			APIKeyCertificateAlias: "ballerina",
			APIKeyIssuer:           "https://localhost:9095/apikey",
			APIKeyAudience:         "http://org.wso2.apimgt/gateway",
			ValidateAllowedAPIs:    false,
		},
	},

	// enable observability of MGW
	ObservabilityEnabled: false,

	// JWT Header Configs
	JwtHeader: "X-JWT-Assertion",

	// JWT Generator Configs
	EnabledJwtGenerator:         "false",
	JwtClaimDialect:             "http://wso2.org/claims",
	JwtSigningAlgorithm:         "SHA256withRSA",
	JwtCertificateAlias:         "ballerina",
	JwtPrivateKeyAlias:          "ballerina",
	JwtTokenExpiry:              900000,
	JwtIssuer:                   "wso2.org/products/am",
	JwtGeneratorImpl:            "org.wso2.micro.gateway.jwt.generator.MGWJWTGeneratorImpl",
	JwtTokenCacheEnable:         "true",
	JwtTokenCacheExpiryTime:     900000,
	JwtTokenCacheCapacity:       10000,
	JwtTokenCacheEvictionFactor: 0.25,
}

// SetApimConfigs sets the MGW configs from APIM configmap
func SetApimConfigs(client *client.Client) error {
	// get data from APIM configmap
	apimConfig := k8s.NewConfMap()
	errApim := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: apimConfName}, apimConfig)

	if errApim != nil {
		if errors.IsNotFound(errApim) {
			logConf.Info("APIM config is not found. Continue with default configs")
		} else {
			logConf.Error(errApim, "Error retrieving APIM configs")
			return errApim
		}
	}

	apimSecret := k8s.NewSecret()
	errorApimSecret := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: apimSecretName}, apimSecret)

	if errorApimSecret != nil {
		if errors.IsNotFound(errorApimSecret) {
			logConf.Info("APIM secret is not found. Continue with default configs")
		} else {
			logConf.Error(errApim, "Error retrieving APIM secret configs")
			return errorApimSecret
		}
	}

	Configs.VerifyHostname = apimConfig.Data[verifyHostnameConst]
	Configs.EnabledGlobalTMEventPublishing = apimConfig.Data[enabledGlobalTMEventPublishingConst]
	Configs.JmsConnectionProvider = apimConfig.Data[jmsConnectionProviderConst]
	Configs.ThrottleEndpoint = apimConfig.Data[throttleEndpointConst]
	Configs.ApimUsername = string(apimSecret.Data["username"])
	Configs.ApimPassword = string(apimSecret.Data["password"])
	Configs.ThrottlingRecieverURL = apimConfig.Data[throttlingRecieverURLConst]
	Configs.ThrottlingAuthURL = apimConfig.Data[throttlingAuthURLConst]
	Configs.EnableRealtimeMessageRetrieval = apimConfig.Data[enableRealtimeMessageRetrievalConst]
	Configs.EnableRequestValidation = apimConfig.Data[enableRequestValidationConst]
	Configs.EnableResponseValidation = apimConfig.Data[enableResponseValidationConst]
	Configs.EnabledEventhub = apimConfig.Data[enabledEventhub]
	Configs.LogLevel = apimConfig.Data[logLevelConst]
	Configs.EnabledAPIKeyIssuer = apimConfig.Data[enabledAPIKeyIssuerConst]
	Configs.APIKeyKeystorePath = apimConfig.Data[apiKeyKeystorePathConst]
	Configs.APIKeyKeystorePassword = apimConfig.Data[apiKeyKeystorePasswordConst]
	Configs.APIKeyIssuerName = apimConfig.Data[apiKeyIssuerNameConst]
	Configs.APIKeyIssuerCertificateAlias = apimConfig.Data[apiKeyIssuerCertificateAliasConst]
	validityTime, err := strconv.Atoi(apimConfig.Data[validityTimeConst])
	if err != nil {
		logConf.Error(err, "Provided validity time is not valid. Using the default validity time")
		Configs.ValidityTime = validityTimeValConst
	} else {
		Configs.ValidityTime = int32(validityTime)
	}
	var apiKeyTokenIssuerAPIs APIKeyTokenAllowedAPIs
	_ = yaml.Unmarshal([]byte(apimConfig.Data[allowedAPIsConst]), &apiKeyTokenIssuerAPIs)
	logConf.Info("API KEY", "apiKeyTokenAPIS", apiKeyTokenIssuerAPIs)
	Configs.APIKeyAllowedAPIs = apiKeyTokenIssuerAPIs
	httpPort, err := strconv.Atoi(apimConfig.Data[httpPortConst])
	if err != nil {
		logConf.Error(err, "Provided http port is not valid. Using the default port")
		Configs.HttpPort = httpPortValConst
	} else {
		Configs.HttpPort = int32(httpPort)
	}
	httpsPort, err := strconv.Atoi(apimConfig.Data[httpsPortConst])
	if err != nil {
		logConf.Error(err, "Provided https port is not valid. Using the default port")
		Configs.HttpsPort = httpsPortValConst
	} else {
		Configs.HttpsPort = int32(httpsPort)
	}
	Configs.JwtHeader = apimConfig.Data[jwtHeaderConst]
	Configs.EnabledJwtGenerator = apimConfig.Data[enabledJwtGeneratorConst]
	Configs.JwtClaimDialect = apimConfig.Data[jwtClaimDialectConst]
	Configs.JwtSigningAlgorithm = apimConfig.Data[jwtSigningAlgorithmConst]
	Configs.JwtCertificateAlias = apimConfig.Data[jwtCertificateAliasConst]
	Configs.JwtPrivateKeyAlias = apimConfig.Data[jwtPrivateKeyAliasConst]
	jwtTokenExpiry, err := strconv.Atoi(apimConfig.Data[jwtTokenExpiryConst])
	if err != nil {
		logConf.Error(err, "Provided JWT token Expiry time is invalid. Using the default JWT token expiry time")
		Configs.JwtTokenExpiry = jwtTokenExpiryValConst
	} else {
		Configs.JwtTokenExpiry = int32(jwtTokenExpiry)
	}
	Configs.JwtIssuer = apimConfig.Data[jwtIssuerConst]
	Configs.JwtGeneratorImpl = apimConfig.Data[jwtGeneratorImplConst]
	Configs.JwtTokenCacheEnable = apimConfig.Data[jwtTokenCacheEnableConst]
	jwtTokenCacheExpiryTime, err := strconv.Atoi(apimConfig.Data[jwtTokenCacheExpiryTimeConst])
	if err != nil {
		logConf.Error(err, "Provided JWT Token cache expiry time is invalid. Using the default JWT token cache"+
			"expiry time")
		Configs.JwtTokenCacheExpiryTime = jwtTokenCacheExpiryTimeValConst
	} else {
		Configs.JwtTokenCacheExpiryTime = int32(jwtTokenCacheExpiryTime)
	}
	var jwtRestrictedClaims []string
	jwtRestrictedClaimsArray := strings.Split(apimConfig.Data[jwtRestrictedClaimsConst], ",")
	for _, element := range jwtRestrictedClaimsArray {
		if element != "" {
			jwtRestrictedClaims = append(jwtRestrictedClaims, element)
		}
	}
	Configs.JwtRestrictedClaims = jwtRestrictedClaims
	var jwtAudience []string
	jwtAudienceArray := strings.Split(apimConfig.Data[jwtAudienceConst], ",")
	for _, element := range jwtAudienceArray {
		if element != "" {
			jwtAudience = append(jwtAudience, element)
		}
	}
	Configs.JwtAudience = jwtAudience
	jwtTokenCacheCapacity, err := strconv.Atoi(apimConfig.Data[jwtTokenCacheCapacityConst])
	if err != nil {
		logConf.Error(err, "Provided JWT Token cache capacity is invalid. Using the default JWT token cache"+
			"capacity")
		Configs.JwtTokenCacheCapacity = jwtTokenCacheCapacityValConst
	} else {
		Configs.JwtTokenCacheCapacity = int32(jwtTokenCacheCapacity)
	}
	jwtTokenCacheEvictionFactor, err := strconv.ParseFloat(apimConfig.Data[jwtTokenCacheEvictionFactorConst], 64)
	if err != nil {
		logConf.Error(err, "Provided JWT Token cache eviction factor is invalid. Using the default JWT token cache"+
			"eviction factor")
		Configs.JwtTokenCacheEvictionFactor = jwtTokenCacheEvictionFactorValConst
	} else {
		Configs.JwtTokenCacheEvictionFactor = jwtTokenCacheEvictionFactor
	}

	return nil
}

// ApplyConfFile render and add the MGW configuration file to cluster
func ApplyConfFile(client *client.Client, userNamespace, apiName string, owner *[]metav1.OwnerReference) error {
	// retrieving the MGW template configmap
	templateConfMap := k8s.NewConfMap()
	errConf := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: mgwConfMustache}, templateConfMap)
	if errConf != nil {
		logConf.Error(errConf, "Error retrieving the MGW template configmap")
		return errConf
	}

	// render final micro-gw-conf file
	templateText := templateConfMap.Data[mgwConfGoTmpl]
	finalConf, errRender := str.RenderTemplate(templateText, Configs)
	if errRender != nil {
		logConf.Error(errRender, "Error rendering the MGW configuration file")
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
