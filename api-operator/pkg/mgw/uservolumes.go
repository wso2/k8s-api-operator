package mgw

import (
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
	"strings"
)

// mgw-deployment-configs
const (
	mgwDeploymentConfigMapName = "mgw-deployment-configs"
	mgwConfigMaps              = "mgwConfigMaps"
	mgwSecrets                 = "mgwSecrets"
)

type DeploymentConfig struct {
	Name          string `yaml:"name"`
	MountLocation string `yaml:"mountLocation"`
	SubPath       string `yaml:"subPath"`
	Namespace     string `yaml:"namespace,omitempty"`
}

var logDeploy = log.Log.WithName("mgw.userDeploymentVolume")

// UserDeploymentVolume returns the deploy volumes and volume mounts with user defined config maps and secrets
func UserDeploymentVolume(client *client.Client, api *wso2v1alpha1.API) ([]corev1.Volume, []corev1.VolumeMount,
	error) {
	var deployVolume []corev1.Volume
	var deployVolumeMount []corev1.VolumeMount
	var mgwDeployVol *v1.Volume
	var mgwDeployMount *v1.VolumeMount
	mgwDeploymentConfMap := k8s.NewConfMap()
	errGetDeploy := k8s.Get(client, types.NamespacedName{Name: mgwDeploymentConfigMapName, Namespace: api.Namespace},
		mgwDeploymentConfMap)
	if errGetDeploy != nil && errors.IsNotFound(errGetDeploy) {
		logDeploy.Info("Get mgw deployment configs", "from namespace", wso2NameSpaceConst)
		//retrieve mgw deployment configs from wso2-system namespace
		err := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: mgwDeploymentConfigMapName},
			mgwDeploymentConfMap)
		if err != nil && !errors.IsNotFound(err) {
			logDeploy.Error(err, "MGW Deployment configs not defined")
			return nil, nil, err
		}
	} else if errGetDeploy != nil {
		logDeploy.Error(errGetDeploy, "Error getting mgw deployment configs from user namespace")
		return nil, nil, errGetDeploy
	}

	var deploymentConfigMaps []DeploymentConfig
	yamlErrDeploymentConfigMaps := yaml.Unmarshal([]byte(mgwDeploymentConfMap.Data[mgwConfigMaps]), &deploymentConfigMaps)
	if yamlErrDeploymentConfigMaps != nil {
		logDeploy.Error(yamlErrDeploymentConfigMaps, "Error marshalling mgw config maps yaml",
			"configmap", mgwDeploymentConfMap)
	}
	var deploymentSecrets []DeploymentConfig
	yamlErrDeploymentSecrets := yaml.Unmarshal([]byte(mgwDeploymentConfMap.Data[mgwSecrets]), &deploymentSecrets)
	if yamlErrDeploymentSecrets != nil {
		logDeploy.Error(yamlErrDeploymentSecrets, "Error marshalling mgw secrets yaml", "configmap",
			mgwDeploymentConfMap)
	}
	// mount the MGW config maps to volume
	for _, deploymentConfigMap := range deploymentConfigMaps {
		if deploymentConfigMap.Namespace == "" {
			mgwConfigMap := k8s.NewConfMap()
			mgwConfigMapErr := k8s.Get(client, types.NamespacedName{Namespace: mgwDeploymentConfMap.Namespace,
				Name: deploymentConfigMap.Name}, mgwConfigMap)
			if mgwConfigMapErr != nil {
				logDeploy.Error(mgwConfigMapErr, "Error Getting the mgw Config map")
			}
			newMgwConfigMap := CopyMgwConfigMap(types.NamespacedName{Namespace: api.Namespace,
				Name: deploymentConfigMap.Name}, mgwConfigMap)
			createConfigMapErr := k8s.Apply(client, newMgwConfigMap)
			if createConfigMapErr != nil {
				logDeploy.Error(createConfigMapErr, "Error Copying mgw config map to user namespace")
			}
			mgwDeployVol, mgwDeployMount = k8s.MgwConfigDirVolumeMount(deploymentConfigMap.Name,
				deploymentConfigMap.MountLocation, deploymentConfigMap.SubPath)
			deployVolume = append(deployVolume, *mgwDeployVol)
			deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
		} else if strings.EqualFold(deploymentConfigMap.Namespace, api.Namespace) {
			mgwDeployVol, mgwDeployMount = k8s.MgwConfigDirVolumeMount(deploymentConfigMap.Name,
				deploymentConfigMap.MountLocation, deploymentConfigMap.SubPath)
			deployVolume = append(deployVolume, *mgwDeployVol)
			deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
		}
	}
	// mount MGW secrets to volume
	for _, deploymentSecret := range deploymentSecrets {
		if deploymentSecret.Namespace == "" {
			mgwSecret := k8s.NewSecret()
			mgwSecretErr := k8s.Get(client, types.NamespacedName{Namespace: mgwDeploymentConfMap.Namespace,
				Name: deploymentSecret.Name}, mgwSecret)
			if mgwSecretErr != nil {
				logDeploy.Error(mgwSecretErr, "Error Getting the mgw Secret")
			}
			newMgwSecret := CopyMgwSecret(types.NamespacedName{Namespace: api.Namespace,
				Name: deploymentSecret.Name}, mgwSecret)
			createSecretErr := k8s.Apply(client, newMgwSecret)
			if createSecretErr != nil {
				logDeploy.Error(createSecretErr, "Error Copying mgw secret to user namespace")
			}
			mgwDeployVol, mgwDeployMount = k8s.MgwSecretVolumeMount(deploymentSecret.Name,
				deploymentSecret.MountLocation,
				deploymentSecret.SubPath)
			deployVolume = append(deployVolume, *mgwDeployVol)
			deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
		} else if strings.EqualFold(deploymentSecret.Namespace, api.Namespace) {
			mgwDeployVol, mgwDeployMount = k8s.MgwSecretVolumeMount(deploymentSecret.Name,
				deploymentSecret.MountLocation,
				deploymentSecret.SubPath)
			deployVolume = append(deployVolume, *mgwDeployVol)
			deployVolumeMount = append(deployVolumeMount, *mgwDeployMount)
		}
	}
	return deployVolume, deployVolumeMount, nil

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
