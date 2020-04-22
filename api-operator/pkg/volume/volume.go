package volume

import corev1 "k8s.io/api/core/v1"

func SecretVolume(secretName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := secretName + "-volume"
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

func ConfigMapVolume(confMapName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := confMapName + "-volume"
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

func EmptyDirVolume(volumeName string, mountPath string) (*corev1.Volume, *corev1.VolumeMount) {
	volName := volumeName + "-volume"
	vol := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	mount := corev1.VolumeMount{
		Name:      volName,
		MountPath: mountPath,
		ReadOnly:  true,
	}
	return &vol, &mount
}
