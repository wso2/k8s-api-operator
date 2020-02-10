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
			Name:      "docker-secret-volume",
			MountPath: "/kaniko/.docker",
			ReadOnly:  true,
		},
	},
	Volumes: []corev1.Volume{
		{
			Name: "docker-secret-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: dockerConfig,
					Items: []corev1.KeyToPath{
						{
							Key:  ".dockerconfigjson",
							Path: "config.json",
						},
					},
				},
			},
		},
	},
}

func dockerHubFunc(repoName string, imgName string) *Config {
	dockerHub.Args = []string{
		fmt.Sprintf("--destination=%s/%s", repoName, imgName),
	}
	return dockerHub
}

func init() {
	addRegistryConfig(DockerHub, dockerHubFunc)
}
