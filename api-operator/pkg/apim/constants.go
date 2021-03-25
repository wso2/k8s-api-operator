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

	publisherAPIImportEndpoint              = "api/am/publisher/v2/apis/import?overwrite=true"
	defaultClientRegistrationEndpointSuffix = "client-registration/v0.17/register"
	defaultApiListEndpointSuffix            = "api/am/publisher/v2/apis"
	defaultTokenEndpoint                    = "oauth2/token"
	importAPIFromSwaggerEndpoint            = "api/am/publisher/v2/apis/import-openapi"
)

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
