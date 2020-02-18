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
	Env              []corev1.EnvVar
	ImagePullSecrets []corev1.LocalObjectReference
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
