package cert

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/volume"
	corev1 "k8s.io/api/core/v1"
	"math/rand"
)

const (
	path = "/usr/wso2/certs/"
)

func Add(cert *corev1.Secret, aliasPrefix string) string {
	// add to cert list
	alias := cert.Name + aliasPrefix
	fileName, _ := maps.OneKey(cert.Data)
	filePath := path + cert.Name + "/" + fileName
	(*kaniko.DocFileProp).Certs[alias] = filePath
	(*kaniko.DocFileProp).CertFound = true

	// add volumes
	name := fmt.Sprintf("%v-%v-cert", alias, rand.Intn(1000))
	certVol := corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: cert.Name,
			},
		},
	}
	certVolMount := corev1.VolumeMount{
		Name:      name,
		MountPath: filePath,
		ReadOnly:  true,
	}
	volume.AddVolume(&certVol, &certVolMount)
	return alias
}
