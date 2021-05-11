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

package vol

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"strings"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// mgw-deployment-configs
const (
	MgwDeploymentConfigMapName = "mgw-deployment-configs"
	mgwConfigMaps              = "mgwConfigMaps"
	mgwSecrets                 = "mgwSecrets"
)

type Context string

func (c Context) isEqual(str string) bool {
	if str == "" {
		return c == DefaultContext
	}
	return strings.EqualFold(string(c), str)
}

const (
	DefaultContext Context = "default"
	KanikoContext  Context = "kaniko"
)

type DeploymentConfig struct {
	Name          string `yaml:"name"`
	MountLocation string `yaml:"mountLocation"`
	SubPath       string `yaml:"subPath"`
	Namespace     string `yaml:"namespace,omitempty"`
	AsEnvVar      bool   `yaml:"asEnvVar,omitempty"`
	Context       string `yaml:"context,omitempty"`
}

var logVol = log.Log.WithName("vol.userDeploymentVolume")

// UserDeploymentVolume returns the deploy volumes and volume mounts with user defined config maps and secrets
// user volumes are returned base on the volume context, if not defined use the default context
func UserDeploymentVolume(client *client.Client, api *wso2v1alpha1.API, volCtx Context) ([]corev1.Volume, []corev1.VolumeMount, []corev1.EnvFromSource,
	error) {
	var deployVolume []corev1.Volume
	var deployVolumeMount []corev1.VolumeMount
	var envFromSources []corev1.EnvFromSource
	var mgwDeployVol *v1.Volume
	var mgwDeployMount *v1.VolumeMount
	mgwDeploymentConfMap := k8s.NewConfMap()
	errGetDeploy := k8s.Get(client, types.NamespacedName{Name: MgwDeploymentConfigMapName, Namespace: api.Namespace},
		mgwDeploymentConfMap)
	if errGetDeploy != nil && errors.IsNotFound(errGetDeploy) {
		logVol.Info("Get mgw deployment configs", "from_namespace", config.SystemNamespace)
		//retrieve mgw deployment configs from wso2-system namespace
		err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: MgwDeploymentConfigMapName},
			mgwDeploymentConfMap)
		if err != nil && !errors.IsNotFound(err) {
			logVol.Error(err, "Error while reading user volumes config")
			return nil, nil, nil, err
		}
	} else if errGetDeploy != nil {
		logVol.Error(errGetDeploy, "Error getting mgw deployment configs from user namespace")
		return nil, nil, nil, errGetDeploy
	}

	var deploymentConfigMaps []DeploymentConfig
	yamlErrDeploymentConfigMaps := yaml.Unmarshal([]byte(mgwDeploymentConfMap.Data[mgwConfigMaps]), &deploymentConfigMaps)
	if yamlErrDeploymentConfigMaps != nil {
		logVol.Error(yamlErrDeploymentConfigMaps, "Error marshalling mgw config maps yaml",
			"configmap", mgwDeploymentConfMap)
		return nil, nil, nil, yamlErrDeploymentConfigMaps
	}
	var deploymentSecrets []DeploymentConfig
	yamlErrDeploymentSecrets := yaml.Unmarshal([]byte(mgwDeploymentConfMap.Data[mgwSecrets]), &deploymentSecrets)
	if yamlErrDeploymentSecrets != nil {
		logVol.Error(yamlErrDeploymentSecrets, "Error marshalling mgw secrets yaml", "configmap",
			mgwDeploymentConfMap)
		return nil, nil, nil, yamlErrDeploymentSecrets
	}
	// mount the MGW config maps to volume
	for _, deploymentConfigMap := range deploymentConfigMaps {
		// if volume context is different then ignore deploymentConfigMap
		if !volCtx.isEqual(deploymentConfigMap.Context) {
			continue
		}

		if deploymentConfigMap.Namespace == "" {
			mgwConfigMap := k8s.NewConfMap()
			mgwConfigMapErr := k8s.Get(client, types.NamespacedName{Namespace: mgwDeploymentConfMap.Namespace,
				Name: deploymentConfigMap.Name}, mgwConfigMap)
			if mgwConfigMapErr != nil {
				logVol.Error(mgwConfigMapErr, "Error Getting the mgw Config map")
				return nil, nil, nil, mgwConfigMapErr
			}
			newMgwConfigMap := CopyMgwConfigMap(types.NamespacedName{Namespace: api.Namespace,
				Name: deploymentConfigMap.Name}, mgwConfigMap)
			createConfigMapErr := k8s.Apply(client, newMgwConfigMap)
			if createConfigMapErr != nil {
				logVol.Error(createConfigMapErr, "Error Copying mgw config map to user namespace")
				return nil, nil, nil, createConfigMapErr
			}

			if deploymentConfigMap.AsEnvVar {
				envFrom := k8s.MgwEnvFromConfigMap(deploymentConfigMap.Name)
				envFromSources = append(envFromSources, *envFrom)
			} else {
				mgwDeployVol, mgwDeployMount = k8s.MgwConfigDirVolumeMount(deploymentConfigMap.Name,
					deploymentConfigMap.MountLocation, deploymentConfigMap.SubPath)
				deployVolume = append(deployVolume, *mgwDeployVol)
				deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
			}
		} else if strings.EqualFold(deploymentConfigMap.Namespace, api.Namespace) {
			if deploymentConfigMap.AsEnvVar {
				envFrom := k8s.MgwEnvFromConfigMap(deploymentConfigMap.Name)
				envFromSources = append(envFromSources, *envFrom)
			} else {
				mgwDeployVol, mgwDeployMount = k8s.MgwConfigDirVolumeMount(deploymentConfigMap.Name,
					deploymentConfigMap.MountLocation, deploymentConfigMap.SubPath)
				deployVolume = append(deployVolume, *mgwDeployVol)
				deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
			}
		}
	}
	// mount MGW secrets to volume
	for _, deploymentSecret := range deploymentSecrets {
		// if volume context is different then ignore deploymentSecret
		if !volCtx.isEqual(deploymentSecret.Context) {
			continue
		}

		if deploymentSecret.Namespace == "" {
			mgwSecret := k8s.NewSecret()
			mgwSecretErr := k8s.Get(client, types.NamespacedName{Namespace: mgwDeploymentConfMap.Namespace,
				Name: deploymentSecret.Name}, mgwSecret)
			if mgwSecretErr != nil {
				logVol.Error(mgwSecretErr, "Error Getting the mgw Secret")
				return nil, nil, nil, mgwSecretErr
			}
			newMgwSecret := CopyMgwSecret(types.NamespacedName{Namespace: api.Namespace,
				Name: deploymentSecret.Name}, mgwSecret)
			createSecretErr := k8s.Apply(client, newMgwSecret)
			if createSecretErr != nil {
				logVol.Error(createSecretErr, "Error Copying mgw secret to user namespace")
				return nil, nil, nil, createSecretErr
			}
			if deploymentSecret.AsEnvVar {
				envFrom := k8s.MgwEnvFromSecret(deploymentSecret.Name)
				envFromSources = append(envFromSources, *envFrom)
			} else {
				mgwDeployVol, mgwDeployMount = k8s.MgwSecretVolumeMount(deploymentSecret.Name,
					deploymentSecret.MountLocation,
					deploymentSecret.SubPath)
				deployVolume = append(deployVolume, *mgwDeployVol)
				deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
			}
		} else if strings.EqualFold(deploymentSecret.Namespace, api.Namespace) {
			if deploymentSecret.AsEnvVar {
				envFrom := k8s.MgwEnvFromSecret(deploymentSecret.Name)
				envFromSources = append(envFromSources, *envFrom)
			} else {
				mgwDeployVol, mgwDeployMount = k8s.MgwSecretVolumeMount(deploymentSecret.Name,
					deploymentSecret.MountLocation,
					deploymentSecret.SubPath)
				deployVolume = append(deployVolume, *mgwDeployVol)
				deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
			}
		}
	}
	return deployVolume, deployVolumeMount, envFromSources, nil

}

// CopyMgwConfigMap returns a copied configMap object with given namespacedName
func CopyMgwConfigMap(namespacedName types.NamespacedName, confMap *corev1.ConfigMap) *corev1.ConfigMap {
	confMap.ObjectMeta = metav1.ObjectMeta{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
	}
	return confMap
}

// CopyMgwSecret returns a copied secret object with given namespacedName
func CopyMgwSecret(namespacedName types.NamespacedName, secret *corev1.Secret) *corev1.Secret {
	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
	}
	return secret
}
