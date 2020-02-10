package registry

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

const AmazonECR Type = "AMAZON_ECR"

var amazonEcr = &Config{
	RegistryType: AmazonECR,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "amazon-cred-helper",
			MountPath: "/kaniko/.docker/",
			ReadOnly:  true,
		},
		{
			Name:      "aws-credentials",
			MountPath: "/kaniko/.docker/",
			ReadOnly:  true,
		},
	},
	Volumes: []corev1.Volume{
		{
			Name: "amazon-cred-helper",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ConfigJsonVolume,
					},
				},
			},
		},
		{
			Name: "aws-credentials",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: AwsCredentialsVolume,
				},
			},
		},
	},
}

func amazonEcrFunc(repoName string, imgName string) *Config {
	// repository = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository"
	// image path = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository/image:v1"
	amazonEcr.ImagePath = fmt.Sprintf("%s/%s", repoName, imgName)
	return amazonEcr
}

func init() {
	addRegistryConfig(AmazonECR, amazonEcrFunc)
}
