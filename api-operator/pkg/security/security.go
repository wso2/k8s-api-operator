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

package security

import (
	"strings"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logSec = log.Log.WithName("security")

type securitySchemeStruct struct {
	SecurityType string             `json:"type"`
	Scheme       string             `json:"scheme,omitempty"`
	Flows        *authorizationCode `json:"flows,omitempty"`
	In           string             `json:"in"`
	Name         string             `json:"name"`
}

type authorizationCode struct {
	AuthorizationCode scopeSet `json:"authorizationCode,omitempty"`
}

type scopeSet struct {
	AuthorizationUrl string            `json:"authorizationUrl"`
	TokenUrl         string            `json:"tokenUrl"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

func Handle(client *client.Client, securityMap map[string][]string, userNameSpace string, secSchemeDefined bool) (map[string]securitySchemeStruct, *[]mgw.JwtTokenConfig, *[]mgw.APIKeyTokenConfig, error) {
	var securityDefinition = make(map[string]securitySchemeStruct)
	//to add multiple certs with alias

	var jwtConfArray []mgw.JwtTokenConfig
	var apiKeyConfArray []mgw.APIKeyTokenConfig
	securityInstance := &wso2v1alpha1.Security{}
	var certificateSecret = k8s.NewSecret()
	for secName, scopeList := range securityMap {
		//retrieve security instances
		errGetSec := k8s.Get(client, types.NamespacedName{Name: secName, Namespace: userNameSpace}, securityInstance)
		if errGetSec != nil && errors.IsNotFound(errGetSec) {
			return securityDefinition, &jwtConfArray, &apiKeyConfArray, errGetSec
		}
		if strings.EqualFold(securityInstance.Spec.Type, oauthConst) {
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				errc := k8s.Get(client, types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
				if errc != nil && errors.IsNotFound(errc) {
					logSec.Info("defined certificate is not found")
					return securityDefinition, &jwtConfArray, &apiKeyConfArray, errc
				} else {
					logSec.Info("defined certificate successfully retrieved")
				}
				//mount certs
				_ = cert.Add(certificateSecret, "security")

				//get the keymanager server URL from the security kind
				mgw.Configs.KeyManagerServerUrl = securityConf.Endpoint
				//fetch credentials from the secret created
				errGetCredentials := SetCredentials(client, oauthConst, types.NamespacedName{Namespace: userNameSpace, Name: securityConf.Credentials})
				if errGetCredentials != nil {
					logSec.Error(errGetCredentials, "Error occurred when retrieving credentials for Oauth")
				} else {
					logSec.Info("Credentials successfully retrieved for security " + secName)
				}
				if !secSchemeDefined {
					//add scopes
					scopes := map[string]string{}
					for _, scopeValue := range scopeList {
						scopes[scopeValue] = "grant " + scopeValue + " access"
					}
					//creating security scheme
					scheme := securitySchemeStruct{
						SecurityType: oauth2Type,
						Flows: &authorizationCode{
							scopeSet{
								authorizationUrl,
								tokenUrl,
								scopes,
							},
						},
					}
					securityDefinition[secName] = scheme
				}
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, jwtConst) {
			logSec.Info("retrieving data for security type JWT")
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				jwtConf := mgw.JwtTokenConfig{}
				//mount certs
				if securityConf.JwksURL != "" {
					jwtConf.JwksPresent = true
					jwtConf.JwksURL = securityConf.JwksURL
				} else {
					errc := k8s.Get(client, types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
					if errc != nil && errors.IsNotFound(errc) {
						logSec.Info("defined certificate is not found")
						return securityDefinition, &jwtConfArray, &apiKeyConfArray, errc
					} else {
						logSec.Info("defined certificate successfully retrieved")
					}
					alias := cert.Add(certificateSecret, "security")
					jwtConf.CertificateAlias = alias
				}
				jwtConf.ValidateSubscription = securityConf.ValidateSubscription

				if securityConf.Issuer != "" {
					jwtConf.Issuer = securityConf.Issuer
				}
				if securityConf.Audience != "" {
					jwtConf.AudiencePresent = true
					jwtConf.Audience = securityConf.Audience
				}

				logSec.Info("certificate issuer", "issuer", jwtConf.Issuer)
				jwtConfArray = append(jwtConfArray, jwtConf)
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, apiKeyConst) {
			logSec.Info("retrieving data for security type APIKey")
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				apiKeyConf := mgw.APIKeyTokenConfig{}
				if !secSchemeDefined {
					//creating security scheme
					scheme := securitySchemeStruct{
						SecurityType: apiKeyConst,
						In:           apiKeyIn,
						Name:         apiKeyName,
					}
					securityDefinition[secName] = scheme
				}
				apiKeyConf.APIKeyCertificateAlias = securityConf.Alias
				apiKeyConf.ValidateAllowedAPIs = securityConf.ValidateAllowedAPIs
				if securityConf.Issuer != "" {
					apiKeyConf.APIKeyIssuer = securityConf.Issuer
				}
				if securityConf.Audience != "" {
					apiKeyConf.APIKeyAudience = securityConf.Audience
				}
				logSec.Info("certificate issuer", "issuer", apiKeyConf.APIKeyIssuer)
				apiKeyConfArray = append(apiKeyConfArray, apiKeyConf)
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, basicSecurityAndScheme) {
			// "existCert = false" for this scenario and do not change the global "existCert" value
			// i.e. if global "existCert" is true, even though the scenario for this swagger is false keep that value as true

			//fetch credentials from the secret created
			errGetCredentials := SetCredentials(client, "Basic",
				types.NamespacedName{Namespace: userNameSpace, Name: securityInstance.Spec.SecurityConfig[0].Credentials})
			if errGetCredentials != nil {
				logSec.Error(errGetCredentials, "Error occurred when retrieving credentials for Basic")
			} else {
				logSec.Info("Credentials successfully retrieved for security " + secName)
			}
			//creating security scheme
			if !secSchemeDefined {
				scheme := securitySchemeStruct{
					SecurityType: basicSecurityType,
					Scheme:       basicSecurityAndScheme,
				}
				securityDefinition[secName] = scheme
			}
		}
	}
	return securityDefinition, &jwtConfArray, &apiKeyConfArray, nil
}
