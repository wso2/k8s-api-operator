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
			MountPath: "/root/.aws/",
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

func amazonEcrFunc(repoName string, imgName string, tag string) *Config {
	// repository = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository"
	// image path = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository:imageName-v1"
	amazonEcr.ImagePath = fmt.Sprintf("%s:%s-%s", repoName, imgName, tag)
	return amazonEcr
}

func init() {
	addRegistryConfig(AmazonECR, amazonEcrFunc)
}
