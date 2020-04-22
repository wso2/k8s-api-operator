package kaniko

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/volume"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	kanikoImgConst = "kanikoImg"
)

func Job(api *wso2v1alpha1.API, controlConfigData map[string]string, kanikoArgs string, owner *[]metav1.OwnerReference) *batchv1.Job {
	rootUserVal := int64(0)
	jobName := api.Name + "-kaniko"
	if api.Spec.UpdateTimeStamp != "" {
		jobName = jobName + "-" + api.Spec.UpdateTimeStamp
	}

	regConfig := registry.GetConfig()

	kanikoImg := controlConfigData[kanikoImgConst]
	args := append([]string{
		"--dockerfile=/usr/wso2/dockerfile/Dockerfile",
		"--context=/usr/wso2/",
		"--destination=" + regConfig.ImagePath,
	}, regConfig.Args...)

	// if kaniko arguments are provided
	// read kaniko arguments and split them as they are read as a single string
	kanikoArguments := strings.Split(kanikoArgs, "\n")
	if kanikoArguments != nil {
		args = append(args, kanikoArguments...)
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       api.Namespace,
			OwnerReferences: *owner,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      api.Name + "-job",
					Namespace: api.Namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         api.Name + "gen-container",
							Image:        kanikoImg,
							VolumeMounts: *volume.JobVolumeMount,
							Args:         args,
							Env:          regConfig.Env,
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUserVal,
					},
					RestartPolicy: "Never",
					Volumes:       *volume.JobVolume,
				},
			},
		},
	}
}
