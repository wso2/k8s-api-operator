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

const Gcr Type = "GCR"
const GoogleSecretEnvVariable = "GOOGLE_APPLICATION_CREDENTIALS"
const SvcAccKeyMountPath = "/kaniko/.gcr/"
const svcAccKeyVolume = "svc-acc-key-volume"

// Google Container Registry Configs

func getGcrConfigFunc(repoName string, imgName string, tag string) *Config {
	return &Config{
		RegistryType: Gcr,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      svcAccKeyVolume,
				MountPath: SvcAccKeyMountPath,
				ReadOnly:  true,
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: svcAccKeyVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: utils.GcrSvcAccKeySecret,
					},
				},
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  GoogleSecretEnvVariable,
				Value: SvcAccKeyMountPath + utils.GcrSvcAccKeyFile,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{Name: utils.GcrPullSecret},
		},
		ImagePath: fmt.Sprintf("gcr.io/%s/%s:%s", repoName, imgName, tag),
	}
}

func init() {
	addRegistryConfig(Gcr, getGcrConfigFunc)
}
