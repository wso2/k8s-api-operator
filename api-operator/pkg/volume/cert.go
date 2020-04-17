package volume

import corev1 "k8s.io/api/core/v1"

func AddCert(cert *corev1.Secret, jobVolumeMount *[]corev1.VolumeMount, jobVolume *[]corev1.Volume) {
	name := CertConfig + "-" + cert.Name
	// check volume already exists
	for _, volume := range *jobVolume {
		if volume.Name == name {
			return
		}
	}

	*jobVolumeMount = append(*jobVolumeMount, corev1.VolumeMount{
		Name:      name,
		MountPath: CertPath + cert.Name,
		ReadOnly:  true,
	})

	*jobVolume = append(*jobVolume, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: cert.Name,
			},
		},
	})
}
