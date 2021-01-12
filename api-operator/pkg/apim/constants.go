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

package apim

const (
	apimConfName                  = "apim-config"
	clientRegistrationSecret      = "ckcs-secret"
	clientIdConst                 = "clientId"
	clientSecretConst             = "clientSecret"
	apimRegistrationEndpointConst = "apimKeymanagerEndpoint"
	apimPublisherEndpointConst    = "apimPublisherEndpoint"
	apimTokenEndpointConst        = "apimTokenEndpoint"
	apimCredentialsConst          = "apimCredentialsSecret"
	skipVerifyConst               = "insecureSkipVerify"

	HeaderAuthorization           = "Authorization"
	HeaderAccept                  = "Accept"
	HeaderContentType             = "Content-Type"
	HeaderConnection              = "Connection"
	HeaderValueApplicationJSON    = "application/json"
	HeaderValueAuthBasicPrefix    = "Basic"
	HeaderValueAuthBearerPrefix   = "Bearer"
	HeaderValueKeepAlive          = "keep-alive"
	HeaderValueXWWWFormUrlEncoded = "application/x-www-form-urlencoded"
	DefaultHttpRequestTimeout     = 10000

	publisherAPIImportEndpoint              = "api/am/publisher/v1/apis/import?overwrite=true"
	defaultClientRegistrationEndpointSuffix = "client-registration/v0.17/register"
	defaultApiListEndpointSuffix            = "api/am/publisher/v1/apis"
	defaultTokenEndpoint                    = "oauth2/token"
	importAPIFromSwaggerEndpoint            = "api/am/publisher/v1/apis/import-openapi"
)

// APIDefinition represents an API artifact in APIM
type APIDefinitionFile struct {
	Type        string           `json:"type,omitempty" yaml:"type,omitempty"`
	ApimVersion string           `json:"version,omitempty" yaml:"version,omitempty"`
	Data        APIDTODefinition `json:"data,omitempty" yaml:"data,omitempty"`
}

// APIDTODefinition represents an APIDTO artifact in APIM
type APIDTODefinition struct {
	ID                           string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name                         string            `json:"name,omitempty" yaml:"name,omitempty"`
	Description                  string            `json:"description,omitempty" yaml:"description,omitempty"`
	Context                      string            `json:"context,omitempty" yaml:"context,omitempty"`
	Version                      string            `json:"version,omitempty" yaml:"version,omitempty"`
	Provider                     string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	LifeCycleStatus              string            `json:"lifeCycleStatus,omitempty" yaml:"lifeCycleStatus,omitempty"`
	WsdlInfo                     interface{}       `json:"wsdlInfo,omitempty" yaml:"wsdlInfo,omitempty"`
	WsdlURL                      string            `json:"wsdlUrl,omitempty" yaml:"wsdlUrl,omitempty"`
	TestKey                      string            `json:"testKey,omitempty" yaml:"testKey,omitempty"`
	ResponseCachingEnabledKey    bool              `json:"responseCachingEnabled,omitempty" yaml:"responseCachingEnabled,omitempty"`
	CacheTimeout                 int               `json:"cacheTimeout,omitempty" yaml:"cacheTimeout,omitempty"`
	DestinationStatsEnabled      string            `json:"destinationStatsEnabled,omitempty" yaml:"destinationStatsEnabled,omitempty"`
	HasThumbnail                 bool              `json:"hasThumbnail,omitempty" yaml:"hasThumbnail,omitempty"`
	IsDefaultVersion             bool              `json:"isDefaultVersion,omitempty" yaml:"isDefaultVersion,omitempty"`
	EnableSchemaValidation       bool              `json:"enableSchemaValidation,omitempty" yaml:"enableSchemaValidation,omitempty"`
	EnableStore                  bool              `json:"enableStore,omitempty" yaml:"enableStore,omitempty"`
	Type                         string            `json:"type,omitempty" yaml:"type,omitempty"`
	Transport                    []string          `json:"transport,omitempty" yaml:"transport,omitempty"`
	Tags                         []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Policies                     []string          `json:"policies,omitempty" yaml:"policies,omitempty"`
	APIThrottlingPolicy          string            `json:"apiThrottlingPolicy,omitempty" yaml:"apiThrottlingPolicy,omitempty"`
	AuthorizationHeader          string            `json:"authorizationHeader,omitempty" yaml:"authorizationHeader,omitempty"`
	SecurityScheme               []string          `json:"securityScheme,omitempty" yaml:"securityScheme,omitempty"`
	MaxTPS                       interface{}       `json:"maxTps,omitempty" yaml:"maxTps,omitempty"`
	Visibility                   string            `json:"visibility,omitempty" yaml:"visibility,omitempty"`
	VisibleRoles                 []string          `json:"visibleRoles,omitempty" yaml:"visibleRoles,omitempty"`
	VisibleTenants               []string          `json:"visibleTenants,omitempty" yaml:"visibleTenants,omitempty"`
	EndpointSecurity             interface{}       `json:"endpointSecurity,omitempty" yaml:"endpointSecurity,omitempty"`
	GatewayEnvironments          []string          `json:"gatewayEnvironments,omitempty" yaml:"gatewayEnvironments,omitempty"`
	DeploymentEnvironments       []interface{}     `json:"deploymentEnvironments,omitempty" yaml:"deploymentEnvironments,omitempty"`
	Labels                       []string          `json:"labels,omitempty" yaml:"labels,omitempty"`
	MediationPolicies            []interface{}     `json:"mediationPolicies,omitempty" yaml:"mediationPolicies,omitempty"`
	SubscriptionAvailability     string            `json:"subscriptionAvailability,omitempty" yaml:"subscriptionAvailability,omitempty"`
	SubscriptionAvailableTenants []string          `json:"subscriptionAvailableTenants,omitempty" yaml:"subscriptionAvailableTenants,omitempty"`
	AdditionalProperties         map[string]string `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Monetization                 interface{}       `json:"monetization,omitempty" yaml:"monetization,omitempty"`
	AccessControl                string            `json:"accessControl,omitempty" yaml:"accessControl,omitempty"`
	AcessControlRoles            []string          `json:"accessControlRoles,omitempty" yaml:"accessControlRoles,omitempty"`
	BusinessInformation          interface{}       `json:"businessInformation,omitempty" yaml:"businessInformation,omitempty"`
	CorsConfiguration            interface{}       `json:"corsConfiguration,omitempty" yaml:"corsConfiguration,omitempty"`
	WorkflowStatus               []string          `json:"workflowStatus,omitempty" yaml:"workflowStatus,omitempty"`
	CreatedTime                  string            `json:"createdTime,omitempty" yaml:"createdTime,omitempty"`
	LastUpdatedTime              string            `json:"lastUpdatedTime,omitempty" yaml:"lastUpdatedTime,omitempty"`
	EndpointConfig               interface{}       `json:"endpointConfig,omitempty" yaml:"endpointConfig,omitempty"`
	EndpointImplementationType   string            `json:"endpointImplementationType,omitempty" yaml:"endpointImplementationType,omitempty"`
	Scopes                       []interface{}     `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	Operations                   []interface{}     `json:"operations,omitempty" yaml:"operations,omitempty"`
	ThreatProtectionPolicies     interface{}       `json:"threatProtectionPolicies,omitempty" yaml:"threatProtectionPolicies,omitempty"`
	Categories                   []string          `json:"categories,omitempty" yaml:"categories,omitempty"`
	KeyManagers                  []string          `json:"keyManagers,omitempty" yaml:"keyManagers,omitempty"`
}

// APIDefinition represents an API artifact in APIM
type APIDefinition struct {
	ID                                 ID                 `json:"id,omitempty" yaml:"id,omitempty"`
	UUID                               string             `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Description                        string             `json:"description,omitempty" yaml:"description,omitempty"`
	Type                               string             `json:"type,omitempty" yaml:"type,omitempty"`
	Context                            string             `json:"context" yaml:"context"`
	ContextTemplate                    string             `json:"contextTemplate,omitempty" yaml:"contextTemplate,omitempty"`
	Tags                               []string           `json:"tags" yaml:"tags,omitempty"`
	Documents                          []interface{}      `json:"documents,omitempty" yaml:"documents,omitempty"`
	LastUpdated                        string             `json:"lastUpdated,omitempty" yaml:"lastUpdated,omitempty"`
	AvailableTiers                     []AvailableTiers   `json:"availableTiers,omitempty" yaml:"availableTiers,omitempty"`
	AvailableSubscriptionLevelPolicies []interface{}      `json:"availableSubscriptionLevelPolicies,omitempty" yaml:"availableSubscriptionLevelPolicies,omitempty"`
	URITemplates                       []URITemplates     `json:"uriTemplates" yaml:"uriTemplates,omitempty"`
	APIHeaderChanged                   bool               `json:"apiHeaderChanged,omitempty" yaml:"apiHeaderChanged,omitempty"`
	APIResourcePatternsChanged         bool               `json:"apiResourcePatternsChanged,omitempty" yaml:"apiResourcePatternsChanged,omitempty"`
	Status                             string             `json:"status,omitempty" yaml:"status,omitempty"`
	TechnicalOwner                     string             `json:"technicalOwner,omitempty" yaml:"technicalOwner,omitempty"`
	TechnicalOwnerEmail                string             `json:"technicalOwnerEmail,omitempty" yaml:"technicalOwnerEmail,omitempty"`
	BusinessOwner                      string             `json:"businessOwner,omitempty" yaml:"businessOwner,omitempty"`
	BusinessOwnerEmail                 string             `json:"businessOwnerEmail,omitempty" yaml:"businessOwnerEmail,omitempty"`
	Visibility                         string             `json:"visibility,omitempty" yaml:"visibility,omitempty"`
	EndpointSecured                    bool               `json:"endpointSecured,omitempty" yaml:"endpointSecured,omitempty"`
	EndpointAuthDigest                 bool               `json:"endpointAuthDigest,omitempty" yaml:"endpointAuthDigest,omitempty"`
	EndpointUTUsername                 string             `json:"endpointUTUsername,omitempty" yaml:"endpointUTUsername,omitempty"`
	Transports                         string             `json:"transports,omitempty" yaml:"transports,omitempty"`
	InSequence                         string             `json:"inSequence,omitempty" yaml:"inSequence,omitempty"`
	OutSequence                        string             `json:"outSequence,omitempty" yaml:"outSequence,omitempty"`
	FaultSequence                      string             `json:"faultSequence,omitempty" yaml:"faultSequence,omitempty"`
	AdvertiseOnly                      bool               `json:"advertiseOnly,omitempty" yaml:"advertiseOnly,omitempty"`
	CorsConfiguration                  *CorsConfiguration `json:"corsConfiguration,omitempty" yaml:"corsConfiguration,omitempty"`
	ProductionUrl                      string             `json:"productionUrl,omitempty" yaml:"productionUrl,omitempty"`
	SandboxUrl                         string             `json:"sandboxUrl,omitempty" yaml:"sandboxUrl,omitempty"`
	EndpointConfig                     *string            `json:"endpointConfig,omitempty" yaml:"endpointConfig,omitempty"`
	ResponseCache                      string             `json:"responseCache,omitempty" yaml:"responseCache,omitempty"`
	CacheTimeout                       int                `json:"cacheTimeout,omitempty" yaml:"cacheTimeout,omitempty"`
	Implementation                     string             `json:"implementation,omitempty" yaml:"implementation,omitempty"`
	AuthorizationHeader                string             `json:"authorizationHeader,omitempty" yaml:"authorizationHeader,omitempty"`
	Scopes                             []interface{}      `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	IsDefaultVersion                   bool               `json:"isDefaultVersion,omitempty" yaml:"isDefaultVersion,omitempty"`
	IsPublishedDefaultVersion          bool               `json:"isPublishedDefaultVersion,omitempty" yaml:"isPublishedDefaultVersion,omitempty"`
	Environments                       []string           `json:"environments,omitempty" yaml:"environments,omitempty"`
	CreatedTime                        string             `json:"createdTime,omitempty" yaml:"createdTime,omitempty"`
	AdditionalProperties               map[string]string  `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	EnvironmentList                    []string           `json:"environmentList,omitempty" yaml:"environmentList,omitempty"`
	APISecurity                        string             `json:"apiSecurity,omitempty" yaml:"apiSecurity,omitempty"`
	AccessControl                      string             `json:"accessControl,omitempty" yaml:"accessControl,omitempty"`
	Rating                             float64            `json:"rating,omitempty" yaml:"rating,omitempty"`
	IsLatest                           bool               `json:"isLatest,omitempty" yaml:"isLatest,omitempty"`
	EnableStore                        bool               `json:"enableStore,omitempty" yaml:"enableStore,omitempty"`
	KeyManagers                        []string           `json:"keyManagers,omitempty" yaml:"keyManagers,omitempty"`
}
type ID struct {
	ProviderName string `json:"providerName" yaml:"providerName"`
	APIName      string `json:"apiName" yaml:"apiName"`
	Version      string `json:"version" yaml:"version"`
}
type AvailableTiers struct {
	Name               string `json:"name,omitempty" yaml:"name,omitempty"`
	DisplayName        string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description        string `json:"description,omitempty" yaml:"description,omitempty"`
	RequestsPerMin     int    `json:"requestsPerMin,omitempty" yaml:"requestsPerMin,omitempty"`
	RequestCount       int    `json:"requestCount,omitempty" yaml:"requestCount,omitempty"`
	UnitTime           int    `json:"unitTime,omitempty" yaml:"unitTime,omitempty"`
	TimeUnit           string `json:"timeUnit,omitempty" yaml:"timeUnit,omitempty"`
	TierPlan           string `json:"tierPlan,omitempty" yaml:"tierPlan,omitempty"`
	StopOnQuotaReached bool   `json:"stopOnQuotaReached,omitempty" yaml:"stopOnQuotaReached,omitempty"`
}
type Scopes struct {
	Key         string `json:"key,omitempty" yaml:"key,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Roles       string `json:"roles,omitempty" yaml:"roles,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	ID          int    `json:"id,omitempty" yaml:"id,omitempty"`
}
type MediationScripts struct {
}
type URITemplates struct {
	URITemplate          string            `json:"uriTemplate,omitempty" yaml:"uriTemplate,omitempty"`
	HTTPVerb             string            `json:"httpVerb,omitempty" yaml:"httpVerb,omitempty"`
	AuthType             string            `json:"authType,omitempty" yaml:"authType,omitempty"`
	HTTPVerbs            []string          `json:"httpVerbs,omitempty" yaml:"httpVerbs,omitempty"`
	AuthTypes            []string          `json:"authTypes,omitempty" yaml:"authTypes,omitempty"`
	ThrottlingConditions []interface{}     `json:"throttlingConditions,omitempty" yaml:"throttlingConditions,omitempty"`
	ThrottlingTier       string            `json:"throttlingTier,omitempty" yaml:"throttlingTier,omitempty"`
	ThrottlingTiers      []string          `json:"throttlingTiers,omitempty" yaml:"throttlingTiers,omitempty"`
	MediationScript      string            `json:"mediationScript,omitempty" yaml:"mediationScript,omitempty"`
	Scopes               []*Scopes         `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	MediationScripts     *MediationScripts `json:"mediationScripts,omitempty" yaml:"mediationScripts,omitempty"`
}
type CorsConfiguration struct {
	CorsConfigurationEnabled      bool     `json:"corsConfigurationEnabled,omitempty" yaml:"corsConfigurationEnabled,omitempty"`
	AccessControlAllowOrigins     []string `json:"accessControlAllowOrigins,omitempty" yaml:"accessControlAllowOrigins,omitempty"`
	AccessControlAllowCredentials bool     `json:"accessControlAllowCredentials,omitempty" yaml:"accessControlAllowCredentials,omitempty"`
	AccessControlAllowHeaders     []string `json:"accessControlAllowHeaders,omitempty" yaml:"accessControlAllowHeaders,omitempty"`
	AccessControlAllowMethods     []string `json:"accessControlAllowMethods,omitempty" yaml:"accessControlAllowMethods,omitempty"`
}

type API struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Context         string `json:"context"`
	Version         string `json:"version"`
	Provider        string `json:"provider"`
	LifeCycleStatus string `json:"lifeCycleStatus"`
}

type APIListResponse struct {
	Count int32 `json:"count"`
	List  []API `json:"list"`
}

type RESTConfig struct {
	KeyManagerEndpoint    string
	PublisherEndpoint     string
	TokenEndpoint         string
	CredentialsSecretName string
	SkipVerification      bool
}
