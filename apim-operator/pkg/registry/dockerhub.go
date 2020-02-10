package registry

import (
	"fmt"
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
					SecretName: ConfigJsonVolume,
					Items: []corev1.KeyToPath{
						{
							Key:  DockerConfigKeyConst,
							Path: "config.json",
						},
					},
				},
			},
		},
	},
}

func dockerHubFunc(repoName string, imgName string) *Config {
	dockerHub.ImagePath = fmt.Sprintf("docker.io/%s/%s", repoName, imgName)
	return dockerHub
}

func init() {
	addRegistryConfig(DockerHub, dockerHubFunc)
}
