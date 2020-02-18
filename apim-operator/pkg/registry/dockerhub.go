package registry

import (
	"fmt"
	"github.com/wso2/k8s-apim-operator/apim-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
)

const DockerHub Type = "DOCKER_HUB"

var dockerHub = &Config{
	RegistryType: DockerHub,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "reg-secret-volume",
			MountPath: "/kaniko/.docker/",
			ReadOnly:  true,
		},
	},
	Volumes: []corev1.Volume{
		{
			Name: "reg-secret-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: utils.ConfigJsonVolume,
					Items: []corev1.KeyToPath{
						{
							Key:  utils.DockerConfigKeyConst,
							Path: "config.json",
						},
					},
				},
			},
		},
	},
	ImagePullSecrets: []corev1.LocalObjectReference{
		{Name: utils.ConfigJsonVolume},
	},
}

func dockerHubFunc(repoName string, imgName string, tag string) *Config {
	dockerHub.ImagePath = fmt.Sprintf("%s/%s:%s", repoName, imgName, tag)
	return dockerHub
}

func init() {
	addRegistryConfig(DockerHub, dockerHubFunc)
}
