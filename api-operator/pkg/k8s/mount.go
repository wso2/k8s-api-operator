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

package k8s

import corev1 "k8s.io/api/core/v1"

func SecretVolumeMount(secretName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := secretName + "-vol"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		ReadOnly:  true,
	}

	return &vol, &mount
}

func ConfigMapVolumeMount(confMapName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := confMapName + "-vol"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: confMapName,
				},
			},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		ReadOnly:  true,
	}
	return &vol, &mount
}

func EmptyDirVolumeMount(volumeName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := volumeName + "-vol"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		ReadOnly:  false,
	}
	return &vol, &mount
}

func MgwConfigDirVolumeMount(confMapName string, mountPath string, subPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := confMapName + "-vol"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: confMapName,
				},
			},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		SubPath:   subPath,
		ReadOnly:  false,
	}
	return &vol, &mount
}

func MgwSecretVolumeMount(secretName string, mountPath string, subPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := secretName + "-vol"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		SubPath:   subPath,
		ReadOnly:  false,
	}

	return &vol, &mount
}

func MgwEnvFromConfigMap(name string) *corev1.EnvFromSource {
	return &corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
}

func MgwEnvFromSecret(name string) *corev1.EnvFromSource {
	return &corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
}
