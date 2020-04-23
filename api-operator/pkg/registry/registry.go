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

package registry

import (
	"encoding/json"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("registry")

const wso2NameSpaceConst = "wso2-system"

type Type string

// Image defines a docker image
type Image struct {
	RegistryType   Type   // Type of the registry
	RepositoryName string // Name of the Docker repository
	Name           string // Image name
	Tag            string // Image tag
}

// Config defines the registry specific configurations
type Config struct {
	RegistryType     Type                          // Type of the registry
	ImagePath        string                        // Full image path to be pushed by the Kaniko Job
	VolumeMounts     []corev1.VolumeMount          // VolumeMounts for the pod that runs Kaniko Job
	Volumes          []corev1.Volume               // Volumes to be mounted for the pod that runs Kaniko Job
	Env              []corev1.EnvVar               // Environment variables to be set in the pod that runs Kaniko Job
	Args             []string                      // Args to be passed to the Kaniko Job
	ImagePullSecrets []corev1.LocalObjectReference // Secrets for the pod which runs the final micro-gateway setup
	IsImageExist     func(config *Config, auth utils.RegAuth, image string,
		tag string) (bool, error) // Function to check the already existence of the image
}

// registry details
var dockerImage Image
var registryConfigs = map[Type]func(repoName string, imgName string, tag string) *Config{}

// SetRegistry sets the registry type, repository and image
func SetRegistry(client *client.Client, namespace string, img Image) error {
	logger.Info("Setting registry type", "image", img)
	dockerImage = img

	return copyConfigVolumes(client, namespace)
}

// GetConfig returns the registry config
func GetConfig() *Config {
	return registryConfigs[dockerImage.RegistryType](dockerImage.RepositoryName, dockerImage.Name, dockerImage.Tag)
}

// IsRegistryType validates the given regType is a valid registry type
func IsRegistryType(regType string) bool {
	_, ok := registryConfigs[Type(regType)]
	return ok
}

// IsImageExist checks if the image exists in the given registry using the secret in the user-namespace
func IsImageExist(client *client.Client) (bool, error) {
	var registryUrl string
	var username string
	var password string

	type Auth struct {
		Auths map[string]struct {
			Auth     string `json:"auth"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auths"`
	}

	// checks if the secret is available
	logger.Info("Getting Docker credentials secret")
	dockerConfigSecret := k8s.NewSecret()
	err := k8s.Get(client, types.NamespacedName{Name: utils.DockerRegCredSecret, Namespace: wso2NameSpaceConst}, dockerConfigSecret)
	if err == nil && errors.IsNotFound(err) {
		logger.Info("Docker credentials secret is not found", "secret-name", utils.DockerRegCredSecret, "namespace", wso2NameSpaceConst)
	} else if err != nil {
		authsJsonString := dockerConfigSecret.Data[utils.DockerConfigKeyConst]
		auth := Auth{}
		err := json.Unmarshal([]byte(authsJsonString), &auth)
		if err != nil {
			logger.Info("Error unmarshal data of docker credential auth")
		}

		registryUrl, err = maps.OneKey(auth.Auths)
		if err != nil {
			logger.Error(err, "Error in docker config secret", "secret", dockerConfigSecret)
			return false, err
		}
		username = auth.Auths[registryUrl].Username
		password = auth.Auths[registryUrl].Password
	}

	config := GetConfig()
	imageCheckFunc := config.IsImageExist
	image := fmt.Sprintf("%s/%s", dockerImage.RepositoryName, dockerImage.Name)
	regAuth := utils.RegAuth{RegistryUrl: registryUrl, Username: username, Password: password}

	if imageCheckFunc == nil {
		return utils.IsImageExists(regAuth, image, dockerImage.Tag)
	}

	return imageCheckFunc(config, regAuth, image, dockerImage.Tag)
}

// copyConfigVolumes copy the configured secrets and config maps to user's namespace
func copyConfigVolumes(client *client.Client, namespace string) error {
	config := GetConfig()
	for _, volume := range config.Volumes {
		var fromObj runtime.Object
		var name string

		if volume.Secret != nil {
			name = volume.Secret.SecretName
			fromObj = k8s.NewSecret()
		}
		if volume.ConfigMap != nil {
			name = volume.ConfigMap.Name
			fromObj = k8s.NewConfMap()
		}

		fromNsName := types.NamespacedName{Namespace: wso2NameSpaceConst, Name: name}
		if err := k8s.Get(client, fromNsName, fromObj); err != nil {
			return err
		}

		toObj := fromObj
		toObj.(metav1.Object).SetNamespace(namespace)
		toObjMeta := toObj.(metav1.Object)
		toObjMeta.SetResourceVersion("")
		//newObjMeta := metav1.ObjectMeta{Namespace:namespace, Name:name}

		if err := k8s.Apply(client, toObj); err != nil {
			return err
		}
	}

	return nil
}

func addRegistryConfig(regType Type, getConfigFunc func(repoName string, imgName string, tag string) *Config) {
	if registryConfigs[regType] != nil {
		logger.Error(nil, "Duplicate registry types", "type", regType)
	} else {
		registryConfigs[regType] = getConfigFunc
	}
}
