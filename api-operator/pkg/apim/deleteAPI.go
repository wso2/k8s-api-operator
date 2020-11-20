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
	"strings"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logDelete = log.Log.WithName("apim.delete")

func DeleteImportedAPI(client *client.Client, instance *wso2v1alpha1.API) error {
	apimConfig, errInput := getRESTAPIConfigs(client)
	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logDelete.Info("APIM config is not found. Continue with default configs")
			return errInput
		} else {
			logDelete.Error(errInput, "Error retrieving APIM configs")
			return errInput
		}
	}

	kmEndpoint := apimConfig.KeyManagerEndpoint
	publisherEndpoint := apimConfig.PublisherEndpoint
	tokenEndpoint := apimConfig.TokenEndpoint
	credSecret := apimConfig.CredentialsSecretName

	insecure = apimConfig.SkipVerification

	if strings.EqualFold(tokenEndpoint, "") {
		tokenEndpoint = kmEndpoint + "/" + defaultTokenEndpoint
		logDelete.Info("Token endpoint not defined. Using keymanager endpoint.", "tokenEndpoint", tokenEndpoint)
	}

	accessToken, errToken := getAccessToken(client, tokenEndpoint, kmEndpoint, credSecret)
	if errToken != nil {
		return errToken
	}

	//itterate throught all API definition.
	for _, configMapName := range instance.Spec.Definition.SwaggerConfigmapNames {
		inputConf := k8s.NewConfMap()
		errInput := k8s.Get(client, types.NamespacedName{Namespace: instance.Namespace, Name: configMapName}, inputConf)

		if errInput != nil {
			if errors.IsNotFound(errInput) {
				logDelete.Info("API project or swagger not found")
				return errInput
			} else {
				logDelete.Error(errInput, "Error retrieving API configs to import")
				return errInput
			}
		}

		if inputConf.BinaryData != nil {
			deleteErr := deleteAPIFromProject(inputConf, accessToken, publisherEndpoint)
			if deleteErr != nil {
				logDelete.Error(deleteErr, "Error when deleting the API using zip")
				return deleteErr
			}
		} else {
			deleteErr := deleteAPIFromSwagger(inputConf, accessToken, publisherEndpoint)
			if deleteErr != nil {
				logDelete.Error(deleteErr, "Error when deleting the API using swagger")
				return deleteErr
			}
		}
	}
	return nil
}

func deleteAPIFromProject(config *corev1.ConfigMap, token string, endpoint string) error {
	zipFileName, errZip := maps.OneKey(config.BinaryData)
	if errZip != nil {
		return errZip
	}
	zippedData := config.BinaryData[zipFileName]

	tmpPath, err := getTempPathOfExtractedArchive(zippedData)
	if err != nil {
		logDelete.Error(err, "Error while getting extracted temporary directory")
		return err
	}

	// Get API info
	apiInfo, err := getAPIDefinition(tmpPath)
	if err != nil {
		logDelete.Error(err, "Error while getting API definition")
		return err
	}

	// checks whether the API exists in APIM
	apiId, err := getAPIId(token, endpoint+"/"+defaultApiListEndpointSuffix, apiInfo.ID.APIName, apiInfo.ID.Version)
	if err != nil {
		return err
	}

	deleteErr := deleteAPIById(endpoint, apiId, token)
	if deleteErr != nil {
		logDelete.Error(deleteErr, "Error when deleting the API from APIM")
	}

	return nil
}

func deleteAPIFromSwagger(config *corev1.ConfigMap, token string, endpoint string) error {
	swaggerFileName, errSwagger := maps.OneKey(config.Data)
	if errSwagger != nil {
		logImport.Error(errSwagger, "Error in the swagger configmap data", "data", config.Data)
		return errSwagger
	}
	swaggerData := config.Data[swaggerFileName]

	_, name, version, err := getAdditionalProperties(swaggerData)
	if err != nil {
		logImport.Error(err, "Error getting additional data")
		return err
	}

	apiId, err := getAPIId(token, endpoint+"/"+defaultApiListEndpointSuffix, name, version)
	if err != nil {
		return err
	}

	deleteErr := deleteAPIById(endpoint, apiId, token)
	if deleteErr != nil {
		logDelete.Error(deleteErr, "Error when deleting the API from APIM")
		return deleteErr
	}

	return nil
}
