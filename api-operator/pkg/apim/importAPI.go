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
	"github.com/Jeffail/gabs"
	jsoniter "github.com/json-iterator/go"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	specs "github.com/wso2/product-apim-tooling/import-export-cli/specs/params"
	"github.com/wso2/product-apim-tooling/import-export-cli/utils"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logImport = log.Log.WithName("apim.import")
var insecure = true

// ImportAPI imports an API to APIM using either project zip or swagger
func ImportAPI(client *client.Client, api *wso2v1alpha2.API) error {
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

	swaggerCM := k8s.NewConfMap()
	paramsCM := k8s.NewConfMap()
	certsCM := k8s.NewConfMap()
	validateSwaggerCM(client, api, swaggerCM)
	validateParamsCM(client, api, paramsCM)
	validateCertsCM(client, api, certsCM)

	if swaggerCM.BinaryData != nil {
		logImport.Info("Importing API using project zip")
		importErr := importAPIFromZip(swaggerCM, paramsCM, certsCM, accessToken, publisherEndpoint)
		if importErr != nil {
			logImport.Error(importErr, "Error when importing the API using zip")
			return importErr
		}
	} else {
		logImport.Info("Importing API using swagger")
		importErr := importAPIFromSwagger(swaggerCM, accessToken, publisherEndpoint)
		if importErr != nil {
			logImport.Error(importErr, "Error when importing the API using swagger")
			return importErr
		}
	}

	return nil
}

// validateSwaggerCM Validates the Swagger CM
func validateSwaggerCM(client *client.Client, api *wso2v1alpha2.API, config *corev1.ConfigMap) error {

	errInput := k8s.Get(client, types.NamespacedName{
		Namespace: api.Namespace,
		Name:      api.Spec.SwaggerConfigMapName,
	}, config)

	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logImport.Info("API project or swagger not found for API", "API Name", api.Name)
			return errInput
		} else {
			logImport.Error(errInput, "Error retrieving API configs to import for the API", "API Name", api.Name)
			return errInput
		}
	}

	return nil
}

// validateParamsCM Validates the Params CM
func validateParamsCM(client *client.Client, api *wso2v1alpha2.API, config *corev1.ConfigMap) error {

	if api.Spec.ParamsValues == "" {
		return nil
	}

	errInput := k8s.Get(client, types.NamespacedName{
		Namespace: api.Namespace,
		Name:      api.Spec.ParamsValues,
	}, config)

	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logImport.Info("API Params Config Map could not found for the API", "API Name", api.Name)
			return errInput
		} else {
			logImport.Error(errInput, "Error retrieving API Params Config Map for the API", "API Name", api.Name)
			return errInput
		}
	}

	return nil
}

// validateCertsCM Validates the Certs CM
func validateCertsCM(client *client.Client, api *wso2v1alpha2.API, config *corev1.ConfigMap) error {

	if api.Spec.CertsValues == "" {
		return nil
	}

	errInput := k8s.Get(client, types.NamespacedName{
		Namespace: api.Namespace,
		Name:      api.Spec.CertsValues,
	}, config)

	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logImport.Info("API Certs Config Map could not found for the API", "API Name", api.Name)
			return errInput
		} else {
			logImport.Error(errInput, "Error retrieving API Certs Config Map for the API", "API Name", api.Name)
			return errInput
		}
	}

	return nil
}

func importAPIFromZip(config *corev1.ConfigMap, paramsCM *corev1.ConfigMap, certsCM *corev1.ConfigMap, token string,
	endpoint string) error {
	zipFileName, errZip := maps.OneKey(config.BinaryData)
	if errZip != nil {
		return errZip
	}
	zippedData := config.BinaryData[zipFileName]

	var importData string
	importData = string(zippedData)

	// params file is required for certs importing as the cert information is available in params.yaml
	if paramsCM.Name != "" {

		zipContent, handleErr := handleDeploymentValues(zippedData, paramsCM, certsCM)
		if handleErr != nil {
			logImport.Error(handleErr, "Error while handling the param values ")
			return handleErr
		}

		importData = zipContent
	}

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, err := writer.CreateFormField("file")
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(importData))
	if err != nil {
		return err
	}

	requestHeaders := make(map[string]string)
	requestHeaders[HeaderContentType] = writer.FormDataContentType()
	requestHeaders[HeaderAuthorization] = HeaderValueAuthBearerPrefix + " " + token
	requestHeaders[HeaderAccept] = "*/*"
	requestHeaders[HeaderConnection] = HeaderValueKeepAlive
	importEndpoint := endpoint + "/" + publisherAPIImportEndpoint

	resp, err := invokePOSTRequest(importEndpoint, requestHeaders, requestBody.Bytes())
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Unable to upload API. Status:" + resp.Status())
	}

	return nil
}

// handleDeploymentValues Handle Param Values and Cert Values
func handleDeploymentValues(zippedData []byte, paramsCM *corev1.ConfigMap, certsCM *corev1.ConfigMap) (string, error) {

	swaggerDirectory, _ := ioutil.TempDir("", "api-swagger-dir*")

	deploymentDir := filepath.Join(swaggerDirectory, filepath.FromSlash(Deployment))
	err := utils.CreateDirIfNotExist(deploymentDir)
	if err != nil {
		return "", err
	}

	if paramsCM.Name != "" {

		paramsFileName, err := maps.OneKey(paramsCM.Data)
		if err != nil {
			logImport.Error(err, "Error in the params configmap ", "params.yaml", paramsCM.Data)
			return "", err
		}
		paramsData := paramsCM.Data[paramsFileName]

		paramsContent, err := getParamValues(paramsData)
		if err != nil {
			logImport.Error(err, "Error in getting the params values ", "param data", paramsData)
			return "", err
		}

		paramsFile := filepath.Join(swaggerDirectory, filepath.FromSlash(Deployment+"/intermediate_params.yaml"))
		err = ioutil.WriteFile(paramsFile, paramsContent, os.ModePerm)
		if err != nil {
			logImport.Error(err, "Error while writing params values to a file.")
			return "", err
		}

	}

	if certsCM.Name != "" {

		certsDir := filepath.Join(swaggerDirectory, filepath.FromSlash(Deployment+"/"+Certificates))
		err = utils.CreateDirIfNotExist(certsDir)
		if err != nil {
			return "", err
		}

		_, err := maps.ManyKeys(certsCM.Data)
		if err != nil {
			logImport.Error(err, "Error in the certs configmap ", "certs cm", certsCM.Data)
			return "", err
		}

		for fileName, certData := range certsCM.Data {

			certFilePath := filepath.Join(swaggerDirectory, filepath.FromSlash(Deployment+"/"+Certificates+
				"/"+fileName))
			err = ioutil.WriteFile(certFilePath, []byte(certData), os.ModePerm)
			if err != nil {
				logImport.Error(err, "Error while writing cert file content")
				return "", err
			}
		}
	}

	zipFile := filepath.Join(swaggerDirectory, filepath.FromSlash("SourceArchive.zip"))
	err = ioutil.WriteFile(zipFile, zippedData, os.ModePerm)
	if err != nil {
		logImport.Error(err, "Error while writing zip file content")
		return "", err
	}

	swaggerZipFile, err, cleanupFunc := utils.CreateZipFileFromProject(swaggerDirectory, false)

	contents, err := getZipContentAsAString(swaggerZipFile)
	if err != nil {
		return "", err
	}

	//cleanup the temporary artifacts once consuming the zip file
	if cleanupFunc != nil {
		defer cleanupFunc()
	}

	return contents, nil

}

// getZipContentAsAString Gets zip file content as a string value
func getZipContentAsAString(swaggerZipFile string) (string, error) {

	file, err := os.Open(swaggerZipFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	contents := buf.String()

	return contents, nil
}

// getParamValues Returns param value content under the zeroth environment
func getParamValues(paramsData string, ) ([]byte, error) {

	apiParams := specs.ApiParams{}
	unmarshalErr := yaml.Unmarshal([]byte(paramsData), &apiParams)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	configValues := apiParams.Environments[0].Config
	envParamsJson, marshallErr := jsoniter.Marshal(configValues)
	if marshallErr != nil {
		return nil, marshallErr
	}

	//var apiParamsPath string
	configValueContent, err := gabs.ParseJSON(envParamsJson)
	paramsContent, err := utils.JsonToYaml(configValueContent.Bytes())
	if err != nil {
		return nil, err
	}

	return paramsContent, nil
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

	dataString, name, version, err := GetAdditionalProperties(swaggerData)
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
