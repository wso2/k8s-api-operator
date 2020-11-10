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

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logLogin = log.Log.WithName("apim.login")
var clientInfo = make(map[string]string)

type ClientRegistrationResponse struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	ClientName   string `json:"clientName"`
	CallBackURL  string `json:"callBackURL"`
	JsonString   string `json:"jsonString"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// login registers a user with APIM to get clientId and clientSecret
func login(client *client.Client, username string, password string, endpoint string) (cId string, cSecret string, error error) {

	registrationEndpoint := endpoint + "/" + defaultClientRegistrationEndpointSuffix
	clientId, clientSecret, clientErr := getClientIdSecret(username, password, registrationEndpoint)
	if clientErr != nil {
		return "", "", clientErr
	}

	clientInfo[clientIdConst] = clientId
	clientInfo[clientSecretConst] = clientSecret
	clientConfName := types.NamespacedName{Namespace: wso2NameSpaceConst, Name: clientRegistrationSecret}
	clientConfSecret := k8s.NewSecretWith(clientConfName, nil, &clientInfo, nil)

	err := k8s.Apply(client, clientConfSecret)
	if err != nil {
		logLogin.Error(err, "Error creating client info secret")
		return "", "", err
	}

	return clientId, clientSecret, nil
}

// getClientIdSecret returns clientId and clientSecret after registering with APIM
func getClientIdSecret(username string, password string, registrationEndpoint string) (clientID string,
	clientSecret string, err error) {

	requestBody := strings.TrimSpace(`{"callbackUrl": "www.wso2.com",
					"clientName": "operator_api_import",
					"grantType": "password refresh_token",
					"saasApp": true,
					"owner":"` + username + `"
					}`)

	requestHeaders := make(map[string]string)
	requestHeaders[HeaderContentType] = HeaderValueApplicationJSON
	requestHeaders[HeaderAuthorization] = HeaderValueAuthBasicPrefix +
		" " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	resp, err := invokePOSTRequest(registrationEndpoint, requestHeaders, requestBody)

	if err != nil {
		logLogin.Error(err, "Error getting client credentials")
		return "", "", err
	}

	if resp.StatusCode() == http.StatusOK || resp.StatusCode() == http.StatusCreated {
		registrationResponse := ClientRegistrationResponse{}
		jsonErr := json.Unmarshal([]byte(resp.Body()), &registrationResponse)
		if jsonErr != nil {
			logLogin.Error(jsonErr, "Error in client registration response")
			return "", "", jsonErr
		}

		return registrationResponse.ClientID, registrationResponse.ClientSecret, nil
	} else {
		if resp.StatusCode() == http.StatusUnauthorized {
			// 401 Unauthorized
			return "", "",
				fmt.Errorf("authorization failed during client registration process")
		}
		return "", "", fmt.Errorf("Request didn't respond 200 OK: " + resp.Status())
	}
}

// getAccessToken returns an access token to use REST APIs in APIM
func getAccessToken(client *client.Client, tokenEndpoint string, dcrEndpoint string) (accessToken string, error error) {
	var username, password, clientId, clientSecret string

	apimSecret := k8s.NewSecret()
	errToken := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: apimSecretName}, apimSecret)
	if errToken != nil {
		return "", errToken
	}
	username = string(apimSecret.Data["username"])
	password = string(apimSecret.Data["password"])

	if len(clientInfo) != 0 {
		logLogin.Info("Getting clientId and clientSecret from memory")
		clientId = clientInfo[clientIdConst]
		clientSecret = clientInfo[clientSecretConst]
	} else {
		ckcsSecret := k8s.NewSecret()
		errToken = k8s.Get(client,
			types.NamespacedName{Namespace: wso2NameSpaceConst, Name: clientRegistrationSecret}, ckcsSecret)
		if errToken != nil {
			if errors.IsNotFound(errToken) {
				logLogin.Info("Client ID, Client Secret not found. Logging in...")
				clientId, clientSecret, errToken = login(client, username, password, dcrEndpoint)
				if errToken != nil {
					return "", errToken
				}
			} else {
				logLogin.Error(errToken, "Error retrieving CKCS secret")
				return "", errToken
			}
		} else {
			// On a restart, read the ckcs secret and update the in-memory values
			logLogin.Info("Getting clientId and clientSecret from ckcs-secret and setting it to memory")
			clientId = string(ckcsSecret.Data[clientIdConst])
			clientSecret = string(ckcsSecret.Data[clientSecretConst])
			clientInfo[clientIdConst] = clientId
			clientInfo[clientSecretConst] = clientSecret
		}
	}

	requestBody := "grant_type=password&username=" + username + "&password=" + url.QueryEscape(password) +
		"&scope=apim:api_import_export+apim:api_view+apim:api_create+apim:api_delete+apim:api_publish"

	requestHeaders := make(map[string]string)
	requestHeaders[HeaderContentType] = HeaderValueXWWWFormUrlEncoded
	requestHeaders[HeaderAuthorization] = HeaderValueAuthBasicPrefix +
		" " + base64.StdEncoding.EncodeToString([]byte(clientId+":"+clientSecret))
	requestHeaders[HeaderAccept] = HeaderValueApplicationJSON

	resp, err := invokePOSTRequest(tokenEndpoint, requestHeaders, requestBody)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("Unable to get token. Status:" + resp.Status())
	}

	tokenResponse := TokenResponse{}
	tokenErr := json.Unmarshal(resp.Body(), &tokenResponse)
	if tokenErr != nil {
		logLogin.Error(tokenErr, "Error in access token response")
		return "", tokenErr
	}

	return tokenResponse.AccessToken, nil
}
