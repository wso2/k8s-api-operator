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

	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	corev1 "k8s.io/api/core/v1"
)

const DockerHub Type = "DOCKER_HUB"

func getDockerHubConfigFunc(repoName string, imgName string, tag string) *Config {
	// Docker Hub configs
	return &Config{
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
				Name: utils.DockerRegCredVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: utils.DockerRegCredSecret,
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
			{Name: utils.DockerRegCredSecret},
		},
		ImagePath: fmt.Sprintf("%s/%s:%s", repoName, imgName, tag),
	}
}

func init() {
	addRegistryConfig(DockerHub, getDockerHubConfigFunc)
}
