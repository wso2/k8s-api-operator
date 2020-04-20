package volume

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	corev1 "k8s.io/api/core/v1"
	"math/rand"
)

const (
	CertPath = "/usr/wso2/certs/"
)

var (
	CertList *map[string]string
	Exists   bool
)

var initCertList = make(map[string]string)

// InitCerts initialize certs volumes
func InitCerts() {
	CertList = &initCertList
	Exists = false
}

func AddCert(cert *corev1.Secret, aliasPrefix string) string {
	// add to cert list
	alias := cert.Name + aliasPrefix
	fileName, _ := maps.OneKey(cert.Data)
	filePath := CertPath + cert.Name + "/" + fileName
	(*CertList)[alias] = filePath
	Exists = true

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
	addVolume(&certVol, &certVolMount)
	return alias
}
