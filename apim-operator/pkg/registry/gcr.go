package registry

import (
	"fmt"
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
					SecretName: GcrSvcAccKeyVolume,
				},
			},
		},
	},
	Env: []corev1.EnvVar{
		{
			Name:  GoogleSecretEnvVariable,
			Value: "/kaniko/.gcr/" + GcrSvcAccKeyFile,
		},
	},
	ImagePullSecrets: []corev1.LocalObjectReference{
		{Name: ConfigJsonVolume},
	},
}

func gcrFunc(repoName string, imgName string, tag string) *Config {
	gcr.ImagePath = fmt.Sprintf("gcr.io/%s/%s:%s", repoName, imgName, tag)
	return gcr
}

func init() {
	addRegistryConfig(Gcr, gcrFunc)
}
