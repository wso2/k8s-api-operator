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
	"bytes"
	"encoding/base64"
	"fmt"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"gopkg.in/resty.v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

var logDeploy = log.Log.WithName("mgw.envoy.deploy")
var insecureDeploy = true

// Deploy API to Envoy Micro-gateway Adapter using zip file or swagger
func DeployAPItoMgw (client *client.Client, api *wso2v1alpha2.API) error {
	var tempMap map[string]string
	envoyMgwConfig := k8s.NewConfMap()
	errEnvoyMgw := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: envoyMgwConfName},
	envoyMgwConfig)

	if errEnvoyMgw != nil {
		if errors.IsNotFound(errEnvoyMgw) {
			logDeploy.Info("Envoy mgw adapter configs not found. Continue with default configs")
			return errEnvoyMgw
		} else {
			logDeploy.Error(errEnvoyMgw, "Error retrieving Envoy mgw adapter configs")
			return errEnvoyMgw
		}
	}
	inputConf := k8s.NewConfMap()
	errInput := k8s.Get(client, types.NamespacedName{Namespace: api.Namespace,
		Name: api.Spec.SwaggerConfigMapName}, inputConf)
	if errInput != nil {
		if errors.IsNotFound(errInput) {
			logDeploy.Info("API project zip file or swagger not found")
			return errInput
		} else {
			logDeploy.Error(errInput, "Error retrieving API configs to deploy")
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

	resourcePath := mgBasePath + mgDeployResourcePath
	mgwEndpoint := envoyMgwConfig.Data[mgwAdapterHostConst]+resourcePath
	var errInsecureDeploy error
	insecureDeploy, errInsecureDeploy = strconv.ParseBool(envoyMgwConfig.Data[mgwInsecureSkipVerifyConst])
	if errInsecureDeploy != nil {
		return errInsecureDeploy
	}
	if !insecureDeploy {
		errCert := getCert(client, mgwCertSecret)
		if errCert != nil {
			return errCert
		}
	}

	logDeploy.Info("Deploying API to Envoy MGW Adapter")
	return deployAPI(inputConf, authToken, mgwEndpoint, tempMap)

}

func deployAPI(config *corev1.ConfigMap, token string, endpoint string, extraParams map[string]string) error{
	if config.BinaryData != nil {
		logDeploy.Info("Deploying API to mgw using project zip")
		errDeployZip := deployAPIZip(config, token, endpoint, extraParams)
		if errDeployZip != nil {
			logDeploy.Error(errDeployZip, "Error when deploying API to mgw using Project zip")
			return errDeployZip
		}
		return nil

	} else {
		logDeploy.Info("Deploying API to mgw using swagger")
		errDeploySwagger := deployAPISwagger(config, token, endpoint, extraParams)
		if errDeploySwagger != nil {
			logDeploy.Error(errDeploySwagger, "Error when deploying API to mgw using Swagger")
			return errDeploySwagger
		}
		return nil
	}
}

func deployAPIZip(config *corev1.ConfigMap, token string, endpoint string, extraParams map[string]string) error {
	fileName, err := getZipData(config)
	if err != nil {
		return err
	}
	resp, errResp := executeNewFileUploadRequest(endpoint, extraParams, "file",
		fileName, token)
	if errResp != nil {
		return errResp
	}
	if resp.StatusCode() == http.StatusCreated || resp.StatusCode() == http.StatusOK {
		// 201 Created or 200 OK
		fmt.Println("Successfully deployed API.")
		return nil
	} else {
		// We have an HTTP error
		return fmt.Errorf("Unable to upload API. Status:" + resp.Status())
	}
}

func deployAPISwagger(config *corev1.ConfigMap, token string, endpoint string, extraParams map[string]string) error{
	swaggerZipFile, cleanupFunc, errSwaggerData := getSwaggerData(config)
	if errSwaggerData != nil {
		return errSwaggerData
	}

	resp, errResp := executeNewFileUploadRequest(endpoint, extraParams, "file",
		swaggerZipFile, token)

	if errResp != nil {
		return errResp
	}
	//cleanup the temporary artifacts once consuming the zip file
	if cleanupFunc != nil {
		defer cleanupFunc()
	}
	if resp.StatusCode() == http.StatusCreated || resp.StatusCode() == http.StatusOK {
		// 201 Created or 200 OK
		fmt.Println("Successfully deployed API.")
		return nil
	} else {
		// We have an HTTP error
		return fmt.Errorf("Unable to upload API. Status:" + resp.Status())
	}
}

func executeNewFileUploadRequest(uri string, params map[string]string, paramName, path,
	accessToken string) (*resty.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := make(map[string]string)
	headers[HeaderContentType] = writer.FormDataContentType()
	headers[HeaderAuthorization] = HeaderValueAuthBasicPrefix + " " + accessToken
	headers[HeaderAccept] = "*/*"
	headers[HeaderConnection] = HeaderValueKeepAlive
	resp, err := invokePOSTRequestWithBytes(uri, headers, body.Bytes())
	return resp, err
}
