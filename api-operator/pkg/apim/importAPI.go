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
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

var logImport = log.Log.WithName("apim.import")
var insecure = true

// ImportAPI imports an API to APIM using either project zip or swagger
func ImportAPI(client *client.Client, api *wso2v1alpha1.API) error {
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
		logImport.Info("Token endpoint not defined. Using keymanager endpoint.", "tokenEndpoint", tokenEndpoint)
	}
	accessToken, errToken := getAccessToken(client, tokenEndpoint, kmEndpoint, credSecret)
	if errToken != nil {
		return errToken
	}

	//itterate throught all API definition.
	for _, configMapName := range api.Spec.Definition.SwaggerConfigmapNames {
		inputConf := k8s.NewConfMap()
		errInput := k8s.Get(client, types.NamespacedName{Namespace: api.Namespace, Name: configMapName}, inputConf)

		if errInput != nil {
			if errors.IsNotFound(errInput) {
				logImport.Info("API project or swagger not found")
				return errInput
			} else {
				logImport.Error(errInput, "Error retrieving API configs to import")
				return errInput
			}
		}

		if inputConf.BinaryData != nil {
			logImport.Info("Importing API using project zip")
			importErr := importAPIFromZip(inputConf, accessToken, publisherEndpoint)
			if importErr != nil {
				logImport.Error(importErr, "Error when importing the API using zip")
				return importErr
			}
		} else {
			logImport.Info("Importing API using swagger")
			importErr := importAPIFromSwagger(inputConf, accessToken, publisherEndpoint)
			if importErr != nil {
				logImport.Error(importErr, "Error when importing the API using swagger")
				return importErr
			}
		}
	}

	return nil
}

func importAPIFromZip(config *corev1.ConfigMap, token string, endpoint string) error {
	updateAPI := false
	zipFileName, errZip := maps.OneKey(config.BinaryData)
	if errZip != nil {
		return errZip
	}
	zippedData := config.BinaryData[zipFileName]

	tmpPath, err := getTempPathOfExtractedArchive(zippedData)
	if err != nil {
		logImport.Error(err, "Error while getting extracted temporary directory")
		return err
	}

	// Get API info
	apiInfo, err := getAPIDefinition(tmpPath)
	if err != nil {
		logImport.Error(err, "Error while getting API definition")
		return err
	}

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, err := writer.CreateFormField("file")
	if err != nil {
		return err
	}
	_, err = part.Write(zippedData)
	if err != nil {
		return err
	}

	// checks whether the API exists in APIM
	apiId, err := getAPIId(token, endpoint+"/"+defaultApiListEndpointSuffix, apiInfo.ID.APIName, apiInfo.ID.Version)
	if err != nil {
		return err
	}
	if !strings.EqualFold(apiId, "") {
		updateAPI = true
	}

	requestHeaders := make(map[string]string)
	requestHeaders[HeaderContentType] = writer.FormDataContentType()
	requestHeaders[HeaderAuthorization] = HeaderValueAuthBearerPrefix + " " + token
	requestHeaders[HeaderAccept] = "*/*"
	requestHeaders[HeaderConnection] = HeaderValueKeepAlive

	importEndpoint := endpoint + "/" + adminAPIImportEndpoint

	if updateAPI {
		logImport.Info("Updating the existing API using zip", "api",
			apiInfo.ID.APIName+":"+apiInfo.ID.Version)
		importEndpoint += "?overwrite=" + url.QueryEscape(strconv.FormatBool(true))
	}

	resp, err := invokePOSTRequest(importEndpoint, requestHeaders, requestBody.Bytes())
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Unable to upload API. Status:" + resp.Status())
	}

	return nil
}

// importAPIFromSwagger imports an API to APIM when an swagger is provided from a configmap
func importAPIFromSwagger(config *corev1.ConfigMap, token string, endpoint string) error {
	updateAPI := false
	swaggerFileName, errSwagger := maps.OneKey(config.Data)
	if errSwagger != nil {
		logImport.Error(errSwagger, "Error in the swagger configmap data", "data", config.Data)
		return errSwagger
	}
	swaggerData := config.Data[swaggerFileName]

	dataString, name, version, err := getAdditionalProperties(swaggerData)
	if err != nil {
		logImport.Error(err, "Error getting additional data")
		return err
	}

	swaggerFile, dataFile, err := getTempFilesForSwagger(swaggerData, dataString)
	if err != nil {
		logImport.Error(err, "Error creating temporary files for swagger data")
		return err
	}
	defer os.Remove(swaggerFile.Name())
	defer os.Remove(dataFile.Name())

	// use the new artifacts
	finalSwaggerFile, err := os.Open(swaggerFile.Name())
	if err != nil {
		return err
	}
	defer finalSwaggerFile.Close()

	finalDataFile, err := os.Open(dataFile.Name())
	if err != nil {
		return err
	}
	defer finalDataFile.Close()

	// checks whether the API exists in APIM
	apiId, err := getAPIId(token, endpoint+"/"+defaultApiListEndpointSuffix, name, version)
	if err != nil {
		return err
	}
	if !strings.EqualFold(apiId, "") {
		updateAPI = true
	}

	if updateAPI {
		logImport.Info("Updating the existing API using swagger", "api", name+":"+version)
		requestBody := &bytes.Buffer{}
		writer := multipart.NewWriter(requestBody)
		part, err := writer.CreateFormFile("file", swaggerFile.Name())
		if err != nil {
			return err
		}
		_, err = io.Copy(part, finalSwaggerFile)
		writer.Close()

		requestHeaders := make(map[string]string)
		requestHeaders[HeaderContentType] = writer.FormDataContentType()
		requestHeaders[HeaderAuthorization] = HeaderValueAuthBearerPrefix + " " + token
		requestHeaders[HeaderAccept] = "*/*"
		requestHeaders[HeaderConnection] = HeaderValueKeepAlive

		updateEndpoint := endpoint + "/" + defaultApiListEndpointSuffix + "/" + apiId + "/" + "swagger"

		resp, err := invokePUTRequest(updateEndpoint, requestHeaders, requestBody.Bytes())
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("Unable to update API. Status:" + resp.Status())
		}

		return nil
	} else {
		requestBody := &bytes.Buffer{}
		writer := multipart.NewWriter(requestBody)
		part1, err := writer.CreateFormFile("file", swaggerFile.Name())
		if err != nil {
			return err
		}
		_, err = io.Copy(part1, finalSwaggerFile)
		part2, err := writer.CreateFormFile("additionalProperties", dataFile.Name())
		if err != nil {
			return err
		}
		_, err = io.Copy(part2, finalDataFile)
		if err != nil {
			return err
		}
		writer.Close()

		requestHeaders := make(map[string]string)
		requestHeaders[HeaderContentType] = writer.FormDataContentType()
		requestHeaders[HeaderAuthorization] = HeaderValueAuthBearerPrefix + " " + token
		requestHeaders[HeaderAccept] = "*/*"
		requestHeaders[HeaderConnection] = HeaderValueKeepAlive

		resp, err := invokePOSTRequest(endpoint+"/"+importAPIFromSwaggerEndpoint, requestHeaders, requestBody.Bytes())
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusCreated {
			return fmt.Errorf("Unable to import API. Status:" + resp.Status())
		}

		return nil
	}
}
