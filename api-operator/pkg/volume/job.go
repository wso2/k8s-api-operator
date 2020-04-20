package volume

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

const (
	dockerFile         = "dockerfile"
	dockerFileLocation = "/usr/wso2/dockerfile/"

	policyConfigmap    = "policy-configmap"
	policyYamlLocation = "/usr/wso2/policy/"

	mgwConfSecretConst = "mgw-conf"
	mgwConfLocation    = "/usr/wso2/mgwconf/"

	swaggerLocation = "/usr/wso2/swagger/project-%v/"
)

var (
	JobVolume      *[]corev1.Volume
	JobVolumeMount *[]corev1.VolumeMount
	ContainerList  *[]corev1.Container
)

// make capacity as 8 since there are 8 volumes in normal scenario
var (
	initJobVolume      = make([]corev1.Volume, 0, 8)
	initJobVolumeMount = make([]corev1.VolumeMount, 0, 8)
	initContainerList  = make([]corev1.Container, 0, 2)
)

// InitJobVolumes initialize Kaniko job volumes
func InitJobVolumes() {
	JobVolume = &initJobVolume
	JobVolumeMount = &initJobVolumeMount
	ContainerList = &initContainerList
}

func AddDefaultKanikoVolumes(apiName string, swaggerCmNames []string) (*[]corev1.Volume, *[]corev1.VolumeMount) {
	// docker file
	dockerFileVol, dockerFileMount := ConfigMapVolume(apiName+"-"+dockerFile, dockerFileLocation)
	// policy
	policyVol, policyMount := ConfigMapVolume(policyConfigmap, policyYamlLocation)
	// MGW conf file
	mgwConfVol, mgwConfMount := SecretVolume(apiName+"-"+mgwConfSecretConst, mgwConfLocation)

	// swagger file config maps
	swaggerVols := make([]corev1.Volume, 0, len(swaggerCmNames))
	swaggerMounts := make([]corev1.VolumeMount, 0, len(swaggerCmNames))
	for i, swaggerCmName := range swaggerCmNames {
		vol, mount := ConfigMapVolume(swaggerCmName+"-mgw", fmt.Sprintf(swaggerLocation, i+1))
		swaggerVols = append(swaggerVols, *vol)
		swaggerMounts = append(swaggerMounts, *mount)
	}

	vols := []corev1.Volume{*dockerFileVol, *policyVol, *mgwConfVol}
	vols = append(vols, swaggerVols...)
	mounts := []corev1.VolumeMount{*dockerFileMount, *policyMount, *mgwConfMount}
	mounts = append(mounts, swaggerMounts...)
	return &vols, &mounts
}

func addContainers(containers *[]corev1.Container) {
	*ContainerList = append(*ContainerList, *containers...)
}

func addVolume(vols *corev1.Volume, volMounts *corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vols)
	*JobVolumeMount = append(*JobVolumeMount, *volMounts)
}

func addVolumes(vols *[]corev1.Volume, volMounts *[]corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vols...)
	*JobVolumeMount = append(*JobVolumeMount, *volMounts...)
}
