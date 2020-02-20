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

type Config struct {
	RegistryType     Type
	ImagePath        string
	VolumeMounts     []corev1.VolumeMount
	Volumes          []corev1.Volume
	ImagePullSecrets []corev1.LocalObjectReference
	Env              []corev1.EnvVar
	Args             []string
	IsImageExist     func(config *Config, auth utils.RegAuth, image string, tag string, logger logr.Logger) (bool, error)
}

// registry details
var registryType Type
var repositoryName string
var imageName string
var imageTag string

var registryConfigs = map[Type]func(repoName string, imgName string, tag string) *Config{}

func SetRegistry(regType Type, repoName string, imgName string, tag string) {
	log.Info("Setting registry type", "registry-type", regType, "repository", repoName, "image", imgName, "tag", tag)
	registryType = regType
	repositoryName = repoName
	imageName = imgName
	imageTag = tag
}

func GetConfig() *Config {
	return registryConfigs[registryType](repositoryName, imageName, imageTag)
}

func IsRegistryType(regType string) bool {
	_, ok := registryConfigs[Type(regType)]
	return ok
}

func IsImageExists(auth utils.RegAuth, logger logr.Logger) (bool, error) {
	config := GetConfig()
	imageCheckFunc := config.IsImageExist
	image := fmt.Sprintf("%s/%s", repositoryName, imageName)

	if imageCheckFunc == nil {
		return utils.IsImageExists(auth, image, imageTag, logger)
	}

	return imageCheckFunc(config, auth, image, imageTag, logger)
}

func addRegistryConfig(regType Type, configFunc func(repoName string, imgName string, tag string) *Config) {
	if registryConfigs[regType] != nil {
		log.Error(nil, "Duplicate registry types", "type", regType)
	} else {
		registryConfigs[regType] = configFunc
	}
}
