package kaniko

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
)

const (
	swaggerLocation = "/usr/wso2/swagger/project-%v/"
)

var (
	JobVolume      *[]corev1.Volume
	JobVolumeMount *[]corev1.VolumeMount
)

// make capacity as 8 since there are 8 volumes in normal scenario
var (
	initJobVolume      = make([]corev1.Volume, 0, 8)
	initJobVolumeMount = make([]corev1.VolumeMount, 0, 8)
)

// InitJobVolumes initialize Kaniko job volumes
func InitJobVolumes() {
	JobVolume = &initJobVolume
	JobVolumeMount = &initJobVolumeMount
}

func AddDefaultKanikoVolumes(apiName string, swaggerCmNames []string) (*[]corev1.Volume, *[]corev1.VolumeMount) {
	// swagger file config maps
	swaggerVols := make([]corev1.Volume, 0, len(swaggerCmNames))
	swaggerMounts := make([]corev1.VolumeMount, 0, len(swaggerCmNames))
	for i, swaggerCmName := range swaggerCmNames {
		vol, mount := k8s.ConfigMapVolumeMount(swaggerCmName+"-mgw", fmt.Sprintf(swaggerLocation, i+1))
		swaggerVols = append(swaggerVols, *vol)
		swaggerMounts = append(swaggerMounts, *mount)
	}

	return &swaggerVols, &swaggerMounts
}

func AddVolume(vols *corev1.Volume, volMounts *corev1.VolumeMount) {
	*JobVolume = append(*JobVolume, *vols)
	*JobVolumeMount = append(*JobVolumeMount, *volMounts)
}
