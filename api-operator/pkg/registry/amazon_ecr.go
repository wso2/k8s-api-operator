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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/go-logr/logr"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

const AmazonECR Type = "AMAZON_ECR"
const AmazonCredHelperVolume = "amazon-cred-helper"
const AmazonCredHelperMountPath = "/kaniko/.docker/"
const AwsCredFileVolume = "aws-credentials"
const AwsCredFileMountPath = "/root/.aws/"

// Amazon ECR Configs
var amazonEcr = &Config{
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
	IsImageExist: func(config *Config, auth utils.RegAuth, image string, tag string, logger logr.Logger) (bool, error) {
		repoNameSplits := strings.Split(repositoryName, ".")
		awsRegistryId := repoNameSplits[0]
		awsRegion := repoNameSplits[3]
		awsRepoName := strings.Split(repositoryName, "/")[1]

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
			logger.Info("Found the image tag from the AWS ECR repository", "RegistryId", awsRegistryId, "RepositoryName", awsRepoName, "image", imageName, "tag", tag)
			if *id.ImageTag == fmt.Sprintf("%s-%s", imageName, tag) {
				return true, nil
			}
		}

		// not found the image with tag
		return false, nil
	},
}

func getAmazonEcrConfigFunc(repoName string, imgName string, tag string) *Config {
	// repository = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository"
	// image path = <aws_account_id.dkr.ecr.region.amazonaws.com>/repository:imageName-v1"
	amazonEcr.ImagePath = fmt.Sprintf("%s:%s-%s", repoName, imgName, tag)
	return amazonEcr
}

func init() {
	addRegistryConfig(AmazonECR, getAmazonEcrConfigFunc)
}
