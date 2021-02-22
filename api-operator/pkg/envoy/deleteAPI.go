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

package envoy

import (
	"encoding/base64"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apim"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

var logDelete = log.Log.WithName("mgw.envoy.delete")
var insecureDelete = true

// DeleteAPIFromMgw deletes the API from the MGW Adapter
func DeleteAPIFromMgw(client *client.Client, api *wso2v1alpha2.API) error {

	logDelete.Info("Jayanie: DELETED")
	envoyMgwConfig := k8s.NewConfMap()
	errEnvoyMgw := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: envoyMgwConfName},
		envoyMgwConfig)

	if errEnvoyMgw != nil {
		if errors.IsNotFound(errEnvoyMgw) {
			logDelete.Info("Envoy mgw adapter configs not found. Continue with default configs")
			return errEnvoyMgw
		} else {
			logDelete.Error(errEnvoyMgw, "Error retrieving Envoy mgw adapter configs")
			return errEnvoyMgw
		}
	}
	inputConf := k8s.NewConfMap()
	errInput := k8s.Get(client, types.NamespacedName{Namespace: api.Namespace,
		Name: api.Spec.SwaggerConfigMapName}, inputConf)
	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logDelete.Info("API project zip file or swagger not found")
			return errInput
		} else {
			logDelete.Error(errInput, "Error retrieving API configs to delete")
			return errInput
		}
	}

	envoyMgwSecret := k8s.NewSecret()
	errEnvoyMgwSecret := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace,
		Name: envoyMgwSecretName}, envoyMgwSecret)
	if errEnvoyMgwSecret != nil {
		return errEnvoyMgwSecret
	}
	username := string(envoyMgwSecret.Data["username"])
	password := string(envoyMgwSecret.Data["password"])
	mgwCertSecret := string(envoyMgwSecret.Data["mgwCertSecretName"])
	authToken := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	resourcePath := mgBasePath + mgDeleteAPIResourcePath
	mgwEndpoint := envoyMgwConfig.Data[mgwAdapterHostConst] + resourcePath
	var errInsecureDelete error
	insecureDelete, errInsecureDelete = strconv.ParseBool(envoyMgwConfig.Data[mgwInsecureSkipVerifyConst])
	if errInsecureDelete != nil {
		return errInsecureDelete
	}
	if !insecureDelete {
		errCert := getCert(client, mgwCertSecret)
		if errCert != nil {
			return errCert
		}
	}
	logDelete.Info("Deleting API from Envoy MGW Adapter")
	return deleteAPI(inputConf, authToken, mgwEndpoint)
}
func deleteAPI(config *corev1.ConfigMap, token string, endpoint string) error {
	if config.BinaryData != nil {
		logDeploy.Info("Deleting API from mgw using project zip")
		errDeployZip := deleteAPIZip(config, token, endpoint)
		if errDeployZip != nil {
			logDeploy.Error(errDeployZip,
				"Error when deleting API from mgw using Project zip")
			return errDeployZip
		}
		return nil

	} else {
		logDeploy.Info("Deleting API to mgw using swagger")
		errDeploySwagger := deleteAPISwagger(config, token, endpoint)
		if errDeploySwagger != nil {
			logDeploy.Error(errDeploySwagger,
				"Error when deleting API from mgw using Swagger")
			return errDeploySwagger
		}
		return nil
	}
}

func deleteAPIZip(config *corev1.ConfigMap, token string, endpoint string) error {
	zipFileName, errZip := maps.OneKey(config.BinaryData)
	if errZip != nil {
		return errZip
	}
	zippedData := config.BinaryData[zipFileName]

	tmpPath, err := apim.GetTempPathOfExtractedArchive(zippedData)
	if err != nil {
		logDelete.Error(err, "Error while getting extracted temporary directory")
		return err
	}
	// Get API info
	apiInfo, err := apim.GetAPIDefinition(tmpPath)
	if err != nil {
		logDelete.Error(err, "Error while getting API definition")
		return err
	}
	queryParams := make(map[string]string)
	queryParams["apiName"] = apiInfo.Data.Name
	queryParams["version"] = apiInfo.Data.Version

	headers := make(map[string]string)
	headers[HeaderAuthorization] = HeaderValueAuthBasicPrefix + " " + token
	resp, err := invokeDELETERequestWithParams(endpoint, queryParams, headers)

	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	} else if resp.StatusCode() == http.StatusNotFound {
		logDelete.Info("API does not exist")
	}
	return nil
}

func deleteAPISwagger(config *corev1.ConfigMap, token string, endpoint string) error {
	swaggerFileName, errSwagger := maps.OneKey(config.Data)
	if errSwagger != nil {
		logDelete.Error(errSwagger, "Error in the swagger configmap data", "data", config.Data)
		return errSwagger
	}
	swaggerData := config.Data[swaggerFileName]

	_, apiName, apiVersion, err := apim.GetAdditionalProperties(swaggerData)
	if err != nil {
		logDelete.Error(err, "Error getting additional data")
		return err
	}
	queryParams := make(map[string]string)
	queryParams["apiName"] = apiName
	queryParams["version"] = apiVersion

	headers := make(map[string]string)
	headers[HeaderAuthorization] = HeaderValueAuthBasicPrefix + " " + token
	resp, err := invokeDELETERequestWithParams(endpoint, queryParams, headers)

	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	} else if resp.StatusCode() == http.StatusNotFound {
		logDelete.Info("API does not exist")
	}
	return nil
}
