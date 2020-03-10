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
	"fmt"
	"github.com/go-logr/logr"
	"github.com/wso2/k8s-apim-operator/apim-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("registry")

type Type string

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
		tag string, logger logr.Logger) (bool, error) // Function to check the already existence of the image
}

// registry details
var registryType Type
var repositoryName string
var imageName string
var imageTag string

var registryConfigs = map[Type]func(repoName string, imgName string, tag string) *Config{}

// SetRegistry sets the registry type, repository and image
func SetRegistry(regType Type, repoName string, imgName string, tag string) {
	log.Info("Setting registry type", "registry-type", regType, "repository", repoName, "image", imgName, "tag", tag)
	registryType = regType
	repositoryName = repoName
	imageName = imgName
	imageTag = tag
}

// GetConfig returns the registry config
func GetConfig() *Config {
	return registryConfigs[registryType](repositoryName, imageName, imageTag)
}

// IsRegistryType validates the given regType is a valid registry type
func IsRegistryType(regType string) bool {
	_, ok := registryConfigs[Type(regType)]
	return ok
}

// IsImageExists check repository for image and returns true if it exists
func IsImageExists(auth utils.RegAuth, logger logr.Logger) (bool, error) {
	config := GetConfig()
	imageCheckFunc := config.IsImageExist
	image := fmt.Sprintf("%s/%s", repositoryName, imageName)

	if imageCheckFunc == nil {
		return utils.IsImageExists(auth, image, imageTag, logger)
	}

	return imageCheckFunc(config, auth, image, imageTag, logger)
}

func addRegistryConfig(regType Type, getConfigFunc func(repoName string, imgName string, tag string) *Config) {
	if registryConfigs[regType] != nil {
		log.Error(nil, "Duplicate registry types", "type", regType)
	} else {
		registryConfigs[regType] = getConfigFunc
	}
}
