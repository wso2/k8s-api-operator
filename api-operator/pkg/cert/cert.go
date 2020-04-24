package cert

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	corev1 "k8s.io/api/core/v1"
)

const (
	path = "/usr/wso2/certs/"
)

func Add(certSecret *corev1.Secret, aliasPrefix string) string {
	// add to cert list
	alias := fmt.Sprintf("%s-%s", certSecret.Name, aliasPrefix)
	fileName, _ := maps.OneKey(certSecret.Data)
	fileDir := path + certSecret.Name
	filePath := fileDir + "/" + fileName
	(*kaniko.DocFileProp).Certs[alias] = filePath
	(*kaniko.DocFileProp).CertFound = true

	// add volumes
	vol, mount := k8s.SecretVolumeMount(certSecret.Name, fileDir)
	kaniko.AddVolume(vol, mount)
	return alias
}
