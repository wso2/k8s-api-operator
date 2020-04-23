package mgw

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	analyticsLocation = "/home/ballerina/wso2/api-usage-data/"
)

// controller config properties
const (
	readinessProbeInitialDelaySeconds = "readinessProbeInitialDelaySeconds"
	readinessProbePeriodSeconds       = "readinessProbePeriodSeconds"
	livenessProbeInitialDelaySeconds  = "livenessProbeInitialDelaySeconds"
	livenessProbePeriodSeconds        = "livenessProbePeriodSeconds"

	resourceRequestCPU    = "resourceRequestCPU"
	resourceRequestMemory = "resourceRequestMemory"
	resourceLimitCPU      = "resourceLimitCPU"
	resourceLimitMemory   = "resourceLimitMemory"
)

var (
	ContainerList     *[]corev1.Container
	initContainerList = make([]corev1.Container, 0, 2)
)

func InitJobVolumes() {
	ContainerList = &initContainerList
}

func AddContainers(containers *[]corev1.Container) {
	*ContainerList = append(*ContainerList, *containers...)
}

// Deployment returns a MGW deployment for the given API definition
func Deployment(api *wso2v1alpha1.API, controlConfigData map[string]string, owner *[]metav1.OwnerReference) *appsv1.Deployment {

	regConfig := registry.GetConfig()
	labels := map[string]string{"app": api.Name}
	var deployVolume []corev1.Volume
	var deployVolumeMount []corev1.VolumeMount

	liveDelay, _ := strconv.ParseInt(controlConfigData[livenessProbeInitialDelaySeconds], 10, 32)
	livePeriod, _ := strconv.ParseInt(controlConfigData[livenessProbePeriodSeconds], 10, 32)
	readDelay, _ := strconv.ParseInt(controlConfigData[readinessProbeInitialDelaySeconds], 10, 32)
	readPeriod, _ := strconv.ParseInt(controlConfigData[readinessProbePeriodSeconds], 10, 32)
	reps := int32(api.Spec.Replicas)

	resReqCPU := controlConfigData[resourceRequestCPU]
	resReqMemory := controlConfigData[resourceRequestMemory]
	resLimitCPU := controlConfigData[resourceLimitCPU]
	resLimitMemory := controlConfigData[resourceLimitMemory]

	if Configs.AnalyticsEnabled {
		// mounts an empty dir volume to be used when analytics is enabled
		analVol, analMount := k8s.EmptyDirVolumeMount("analytics", analyticsLocation)
		deployVolume = append(deployVolume, *analVol)
		deployVolumeMount = append(deployVolumeMount, *analMount)
	}
	req := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resReqCPU),
		corev1.ResourceMemory: resource.MustParse(resReqMemory),
	}
	lim := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resLimitCPU),
		corev1.ResourceMemory: resource.MustParse(resLimitMemory),
	}
	apiContainer := corev1.Container{
		Name:            "mgw" + api.Name,
		Image:           regConfig.ImagePath,
		ImagePullPolicy: "Always",
		Resources: corev1.ResourceRequirements{
			Requests: req,
			Limits:   lim,
		},
		VolumeMounts: deployVolumeMount,
		Env:          regConfig.Env,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: Configs.HttpPort,
			},
			{
				ContainerPort: Configs.HttpsPort,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpsPort},
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: int32(readDelay),
			PeriodSeconds:       int32(readPeriod),
			TimeoutSeconds:      1,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpsPort},
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: int32(liveDelay),
			PeriodSeconds:       int32(livePeriod),
			TimeoutSeconds:      1,
		},
	}

	*(ContainerList) = append(*(ContainerList), apiContainer)

	deploy := k8s.NewDeployment()
	deploy.ObjectMeta = metav1.ObjectMeta{
		Name:            api.Name,
		Namespace:       api.Namespace,
		Labels:          labels,
		OwnerReferences: *owner,
	}
	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: &reps,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers:       *(ContainerList),
				Volumes:          deployVolume,
				ImagePullSecrets: regConfig.ImagePullSecrets,
			},
		},
	}
	return deploy
}
