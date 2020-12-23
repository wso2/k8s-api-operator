// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package mgw

import (
	"strconv"
	"strings"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	envKeyValSeparator = "="
)

// Deployment returns a MGW deployment for the given API definition
func Deployment(client *client.Client, api *wso2v1alpha1.API, controlConfigData map[string]string,
	owner *[]metav1.OwnerReference, sidecarContainers []corev1.Container, img registry.Image) (*appsv1.Deployment, error) {
	regConfig := registry.GetImageConfig(img)
	labels := map[string]string{"app": api.Name}
	annotations := map[string]string{}
	liveDelay, _ := strconv.ParseInt(controlConfigData[livenessProbeInitialDelaySeconds], 10, 32)
	livePeriod, _ := strconv.ParseInt(controlConfigData[livenessProbePeriodSeconds], 10, 32)
	readDelay, _ := strconv.ParseInt(controlConfigData[readinessProbeInitialDelaySeconds], 10, 32)
	readPeriod, _ := strconv.ParseInt(controlConfigData[readinessProbePeriodSeconds], 10, 32)
	reps := int32(api.Spec.Replicas)

	resReqCPU := controlConfigData[resourceRequestCPU]
	resReqMemory := controlConfigData[resourceRequestMemory]
	resLimitCPU := controlConfigData[resourceLimitCPU]
	resLimitMemory := controlConfigData[resourceLimitMemory]

	// Mount the user specified Config maps and secrets to mgw deploy volume
	deployVolume, deployVolumeMount, envFromSources, errDeploy := UserDeploymentVolume(client, api)

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

	// container ports
	containerPorts := []corev1.ContainerPort{
		{
			ContainerPort: Configs.HttpPort,
		},
		{
			ContainerPort: Configs.HttpsPort,
		},
	}
	// setting observability port
	if Configs.ObservabilityEnabled {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: observabilityPrometheusPort,
		})
		annotations["prometheus.io/port"] = strconv.Itoa(observabilityPrometheusPort)
		annotations["prometheus.io/scrape"] = "true"
	}

	// setting environment variables
	// env from registry configs
	env := regConfig.Env
	// env from API CRD Spec
	for _, variable := range api.Spec.EnvironmentVariables {
		envKeyVal := strings.SplitN(variable, envKeyValSeparator, 2)
		env = append(env, corev1.EnvVar{
			Name:  envKeyVal[0],
			Value: envKeyVal[:2][1],
		})
	}

	// setting container image
	var image string
	if api.Spec.Image != "" {
		image = api.Spec.Image
	} else {
		image = regConfig.ImagePath
	}

	// API container
	apiContainer := corev1.Container{
		Name:            "mgw" + api.Name,
		Image:           image,
		ImagePullPolicy: "Always",
		Resources: corev1.ResourceRequirements{
			Requests: req,
			Limits:   lim,
		},
		VolumeMounts: deployVolumeMount,
		Env:          env,
		EnvFrom:      envFromSources,
		Ports:        containerPorts,
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

	// set hostAliases
	hostAliases := getHostAliases(client)

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
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: corev1.PodSpec{
				HostAliases:      hostAliases,
				Containers:       append(sidecarContainers, apiContainer),
				Volumes:          deployVolume,
				ImagePullSecrets: regConfig.ImagePullSecrets,
			},
		},
	}
	return deploy, errDeploy
}
