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

package kaniko

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logDocker = log.Log.WithName("kaniko.docker")

const (
	truststoreSecretName      = "truststorepass"
	dockerFileTemplate        = "dockerfile-template"
	encodedTruststorePassword = "YmFsbGVyaW5h"
	truststoreSecretData      = "password"
	dockerFile                = "dockerfile"
	dockerFileLocation        = "/usr/wso2/dockerfile/"
	SwaggerLocation           = "/usr/wso2/swagger/project-%v/"
)

// DockerfileProperties represents the type for properties of docker file
type DockerfileProperties struct {
	CertFound             bool
	TruststorePassword    string
	Certs                 map[string]string
	ToolkitImage          string
	RuntimeImage          string
	BalInterceptorsFound  bool
	JavaInterceptorsFound bool
}

// DocFileProp represents the properties of docker file
var DocFileProp *DockerfileProperties

func InitDocFileProp() {
	initDocFileProp := &DockerfileProperties{
		CertFound:             false,
		TruststorePassword:    "",
		Certs:                 map[string]string{},
		ToolkitImage:          "",
		RuntimeImage:          "",
		BalInterceptorsFound:  false,
		JavaInterceptorsFound: false,
	}
	DocFileProp = initDocFileProp
}

// HandleDockerFile render the docker file for Kaniko job and add volumes to the Kaniko job
func HandleDockerFile(client *client.Client, userNamespace, apiName string, owner *[]metav1.OwnerReference) error {
	// get docker file template from system namespace
	dockerFileConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: dockerFileTemplate}, dockerFileConfMap)
	if err != nil {
		logDocker.Error(err, "Error retrieving docker template configmap",
			"configmap", dockerFileTemplate, "namespace", userNamespace, "apiName", apiName)
		return err
	}

	// get file name in configmap
	fileName, err := maps.OneKey(dockerFileConfMap.Data)
	if err != nil {
		logDocker.Error(err, "Error retrieving docker template filename",
			"configmap_data", dockerFileConfMap.Data, "namespace", userNamespace, "apiName", apiName)
		return err
	}

	// set truststore password
	if err := setTruststorePassword(client); err != nil {
		return err
	}

	// get rendered docker file
	renderedDocFile, err := str.RenderTemplate(dockerFileConfMap.Data[fileName], DocFileProp)
	if err != nil {
		return err
	}

	// final configmap is the configmap that contains the rendered docker file
	finalConfMapName := fmt.Sprintf("%s-%s", apiName, dockerFile)
	dockerDataMap := map[string]string{"Dockerfile": renderedDocFile}
	finalConfMap := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNamespace, Name: finalConfMapName}, &dockerDataMap, nil, owner)
	err = k8s.Apply(client, finalConfMap)
	if err != nil {
		return err
	}

	// add to job volumes
	vol, mount := k8s.ConfigMapVolumeMount(apiName+"-"+dockerFile, dockerFileLocation)
	AddVolume(vol, mount)

	return nil
}

// setTruststorePassword sets the truststore password in docker file properties DocFileProp
func setTruststorePassword(client *client.Client) error {
	// get secret if available
	secret := k8s.NewSecret()
	err := k8s.Get(client, types.NamespacedName{Name: truststoreSecretName, Namespace: config.SystemNamespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		encodedPw := encodedTruststorePassword
		decodedPw, err := b64.StdEncoding.DecodeString(encodedPw)
		if err != nil {
			logDocker.Error(err, "Error decoding truststore password")
			return err
		}
		password := string(decodedPw)

		logDocker.Info("Creating a new secret for truststore password")
		trustStoreSecret := k8s.NewSecretWith(types.NamespacedName{
			Namespace: config.SystemNamespace,
			Name:      truststoreSecretName,
		}, &map[string][]byte{
			truststoreSecretData: []byte(encodedPw),
		}, nil, nil)

		errSecret := k8s.Create(client, trustStoreSecret)
		logDocker.Info("Error in creating truststore password and ignore it", "error", errSecret)

		DocFileProp.TruststorePassword = password
		return nil
	}
	//get password from the secret
	encodedPw := string(secret.Data[truststoreSecretData])
	decodedPw, err := b64.StdEncoding.DecodeString(encodedPw)
	if err != nil {
		logDocker.Error(err, "Error decoding truststore password")
	}

	DocFileProp.TruststorePassword = string(decodedPw)
	return nil
}
