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

func AddVolume(vols *corev1.Volume, volMounts *corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vols)
	*JobVolumeMount = append(*JobVolumeMount, *volMounts)
}
