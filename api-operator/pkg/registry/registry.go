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
	"strings"

	operatorConfig "github.com/wso2/k8s-api-operator/api-operator/pkg/config"

	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("registry")

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
	IsImageExist     func(config *Config, auth utils.RegAuth, imageRepository string, imageName string,
		tag string) (bool, error) // Function to check the already existence of the image
}

// registry details
var registryConfigs = map[Type]func(repoName string, imgName string, tag string) *Config{}

// SetRegistry sets the registry type, repository and image
func SetRegistry(client *client.Client, namespace string, img Image) error {
	logger.Info("Setting registry type", "image", img)
	return copyConfigVolumes(client, namespace, img)
}

// GetImageConfig returns the registry config for a specific image
func GetImageConfig(image Image) *Config {
	return registryConfigs[image.RegistryType](image.RepositoryName, image.Name, image.Tag)
}

// IsRegistryType validates the given regType is a valid registry type
func IsRegistryType(regType string) bool {
	_, ok := registryConfigs[Type(regType)]
	return ok
}

// IsImageExist checks if the image exists in the given registry using the secret in the user-namespace
func IsImageExist(client *client.Client, image Image) (bool, error) {
	// Auth represents the pull secret of registries
	// local struct since not used in other places
	type Auth struct {
		Auths map[string]struct {
			Auth     string `json:"auth"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auths"`
	}

	config := GetImageConfig(image)
	// checks if the pull secret is available
	regPullSecret := k8s.NewSecret()
	err := k8s.Get(client, types.NamespacedName{Name: config.ImagePullSecrets[0].Name,
		Namespace: operatorConfig.SystemNamespace}, regPullSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Registry pull secret is not found",
				"secret_name", utils.DockerRegCredSecret, "namespace", operatorConfig.SystemNamespace)
		} else {
			logger.Info("Error retrieving docker credentials secret", "secret_name", utils.DockerRegCredSecret, "namespace", operatorConfig.SystemNamespace)
		}
		return false, err
	}

	auth := Auth{}
	err = json.Unmarshal(regPullSecret.Data[utils.DockerConfigKeyConst], &auth)
	if err != nil {
		logger.Info("Error unmarshal data of docker credential auth")
		return false, err
	}

	registryUrl, err := maps.OneKey(auth.Auths)
	if err != nil {
		logger.Info("Error in docker config secret", "secret", regPullSecret)
		return false, err
	}
	username := auth.Auths[registryUrl].Username
	password := auth.Auths[registryUrl].Password

	registryUrl = str.RemoveVersionTag(registryUrl)
	if !strings.HasPrefix(registryUrl, "https://") {
		registryUrl = "https://" + registryUrl
	}
	regAuth := utils.RegAuth{RegistryUrl: registryUrl, Username: username, Password: password}

	// check registry specific image existence functionality is defined
	// if not use default function
	imageCheckFunc := config.IsImageExist
	if imageCheckFunc == nil {
		return utils.IsImageExists(regAuth, image.RepositoryName, image.Name, image.Tag)
	}
	// otherwise defined function
	return imageCheckFunc(config, regAuth, image.RepositoryName, image.Name, image.Tag)
}

// copyConfigVolumes copy the configured secrets and config maps to user's namespace
// from wso2's system namespace
func copyConfigVolumes(client *client.Client, namespace string, image Image) error {
	logger.Info("Replacing configured secrets and config maps for the registry")
	config := GetImageConfig(image)
	// registry volumes map: name -> runtime object
	var regVolumes = make(map[string]runtime.Object, len(config.Volumes)+len(config.ImagePullSecrets))

	// config volumes
	for _, volume := range config.Volumes {
		if volume.Secret != nil {
			regVolumes[volume.Secret.SecretName] = k8s.NewSecret()
		}
		if volume.ConfigMap != nil {
			regVolumes[volume.Secret.SecretName] = k8s.NewConfMap()
		}
	}

	// pull secrets
	for _, pullSecret := range config.ImagePullSecrets {
		regVolumes[pullSecret.Name] = k8s.NewSecret()
	}

	// get object from wso2's system namespace and
	// creates or replaces volumes in the given namespace
	for name, object := range regVolumes {
		fromNsName := types.NamespacedName{Namespace: operatorConfig.SystemNamespace, Name: name}
		if err := k8s.Get(client, fromNsName, object); err != nil {
			return err
		}

		object.(metav1.Object).SetNamespace(namespace)
		objMeta := object.(metav1.Object)
		objMeta.SetResourceVersion("")
		objMeta.SetUID("")

		if err := k8s.Apply(client, object); err != nil {
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
