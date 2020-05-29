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

package kaniko

import (
	corev1 "k8s.io/api/core/v1"
)

var (
	JobVolume      *[]corev1.Volume
	JobVolumeMount *[]corev1.VolumeMount
)

// InitJobVolumes initialize Kaniko job volumes
func InitJobVolumes() {
	initJobVolume := make([]corev1.Volume, 0, 8)
	initJobVolumeMount := make([]corev1.VolumeMount, 0, 8)
	JobVolume = &initJobVolume
	JobVolumeMount = &initJobVolumeMount
}

func AddVolume(vol *corev1.Volume, volMount *corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vol)
	*JobVolumeMount = append(*JobVolumeMount, *volMount)
}

func AddVolumes(vols *[]corev1.Volume, volMounts *[]corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vols...)
	*JobVolumeMount = append(*JobVolumeMount, *volMounts...)
}
