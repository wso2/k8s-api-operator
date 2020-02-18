package registry

import (
	"fmt"
	"github.com/wso2/k8s-apim-operator/apim-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
)

const Gcr Type = "GCR"
const GoogleSecretEnvVariable = "GOOGLE_APPLICATION_CREDENTIALS"

var gcr = &Config{
	RegistryType: Gcr,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "svc-acc-key-volume",
			MountPath: "/kaniko/.gcr/",
			ReadOnly:  true,
		},
	},
	Volumes: []corev1.Volume{
		{
			Name: "svc-acc-key-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: utils.GcrSvcAccKeyVolume,
				},
			},
		},
	},
	Env: []corev1.EnvVar{
		{
			Name:  GoogleSecretEnvVariable,
			Value: "/kaniko/.gcr/" + utils.GcrSvcAccKeyFile,
		},
	},
	ImagePullSecrets: []corev1.LocalObjectReference{
		{Name: utils.ConfigJsonVolume},
	},
}

func gcrFunc(repoName string, imgName string, tag string) *Config {
	gcr.ImagePath = fmt.Sprintf("gcr.io/%s/%s:%s", repoName, imgName, tag)
	return gcr
}

func init() {
	addRegistryConfig(Gcr, gcrFunc)
}
