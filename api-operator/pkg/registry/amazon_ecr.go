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

package registry

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
)

const AmazonECR Type = "AMAZON_ECR"
const AmazonCredHelperVolume = "amazon-cred-helper"
const AmazonCredHelperMountPath = "/kaniko/.docker/"
const AwsCredFileVolume = "aws-credentials"
const AwsCredFileMountPath = "/root/.aws/"

func getAmazonEcrConfigFunc(repoName string, imgName string, tag string) *Config {
	// Amazon ECR Configs
	return &Config{
		RegistryType: AmazonECR,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      AmazonCredHelperVolume,
				MountPath: AmazonCredHelperMountPath,
				ReadOnly:  true,
			},
			{
				Name:      AwsCredFileVolume,
				MountPath: AwsCredFileMountPath,
				ReadOnly:  true,
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: AmazonCredHelperVolume,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: utils.AmazonCredHelperConfMap,
						},
					},
				},
			},
			{
				Name: AwsCredFileVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: utils.AwsCredentialsSecret,
					},
				},
			},
		},
		IsImageExist: func(config *Config, auth utils.RegAuth, imageRepository string, imageName string, tag string) (bool, error) {
			repoNameSplits := strings.Split(imageRepository, ".")
			awsRegistryId := repoNameSplits[0]
			awsRegion := repoNameSplits[3]
			awsRepoName := strings.Split(imageRepository, "/")[1]

			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(awsRegion)},
			)
			if err != nil {
				logger.Error(err, "Error creating aws session")
				return false, err
			}

			svc := ecr.New(sess)
			images, err := svc.ListImages(&ecr.ListImagesInput{
				RegistryId:     &awsRegistryId,
				RepositoryName: &awsRepoName,
			})
			if err != nil {
				logger.Error(err, "Error getting list of images in AWS ECR repository", "RegistryId", awsRegistryId, "RepositoryName", awsRepoName)
				return false, err
			}

			for _, id := range images.ImageIds {
				// found the image with tag
				// untagged images 'id.ImageTag' returns nil; check nil before accessing the pointer
				if id.ImageTag != nil && *id.ImageTag == fmt.Sprintf("%s-%s", imageName, tag) {
					logger.Info("Found the image tag from the AWS ECR repository",
						"RegistryId", awsRegistryId, "RepositoryName", awsRepoName, "image", imageName, "tag", tag)
					return true, nil
				}
			}

			// not found the image with tag
			return false, nil
		},
		// repository = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository"
		// image path = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository:imageName-v1"
		ImagePath: fmt.Sprintf("%s:%s-%s", repoName, imgName, tag),
	}
}

func init() {
	addRegistryConfig(AmazonECR, getAmazonEcrConfigFunc)
}
