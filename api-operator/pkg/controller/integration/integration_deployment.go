/*
 * Copyright (c) 2021 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package integration

import (
	"errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

// deploymentForIntegration returns a integration Deployment object
func (r *ReconcileIntegration) deploymentForIntegration(config EIConfigNew) *appsv1.Deployment {

	var m = config.integration

	//set HTTP and HTTPS ports for as container ports
	exposePorts := []corev1.ContainerPort{
		{
			ContainerPort: m.Spec.Expose.PassthroPort,
		},
	}

	// check inbound endpoint port is exist and append to the container port
	for _, port := range m.Spec.Expose.InboundPorts {
		exposePorts = append(
			exposePorts,
			corev1.ContainerPort{
				ContainerPort: port,
			},
		)
	}

	// check ImagePullPolicy has given with the integration
	var imageSecrets []corev1.LocalObjectReference
	if m.Spec.ImagePullSecret != "" {
		imageSecrets = append(
			imageSecrets,
			corev1.LocalObjectReference{
				Name: m.Spec.ImagePullSecret,
			},
		)
	}

	labels := labelsForIntegration(m.Name)

	replicas := m.Spec.DeploySpec.MinReplicas

	request := corev1.ResourceList{}
	if m.Spec.DeploySpec.ReqCpu != "" {
		request[corev1.ResourceCPU] = resource.MustParse(m.Spec.DeploySpec.ReqCpu)
	}
	if m.Spec.DeploySpec.ReqMemory != "" {
		request[corev1.ResourceMemory] = resource.MustParse(m.Spec.DeploySpec.ReqMemory)
	}

	limit := corev1.ResourceList{}
	if m.Spec.DeploySpec.LimitCpu != "" {
		limit[corev1.ResourceCPU] = resource.MustParse(m.Spec.DeploySpec.LimitCpu)
	}
	if m.Spec.DeploySpec.MemoryLimit != "" {
		limit[corev1.ResourceMemory] = resource.MustParse(m.Spec.DeploySpec.MemoryLimit)
	}

	livenessProbe, _ := getLivenessProbe(config)
	readinessProbe, _ := getReadinessProbe(config)

	volMounts := make([]corev1.VolumeMount, 0)
	volumes := make([]corev1.Volume, 0)
	for _, configMapDetail := range m.Spec.DeploySpec.ConfigMapDetails {
		volumes = append(
			volumes,
			corev1.Volume{
				Name: "volume-" + configMapDetail.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapDetail.Name,
						},
					},
				},
			},
		)

		volMounts = append(
			volMounts,
			corev1.VolumeMount{
				Name:      "volume-" + configMapDetail.Name,
				MountPath: configMapDetail.MountPath + "/" + configMapDetail.FileName,
				SubPath:   configMapDetail.FileName,
			},
		)
	}
	var imagePullPolicy corev1.PullPolicy

	switch pullPolicy := m.Spec.DeploySpec.PullPolicy; pullPolicy {
	case "Never":
		imagePullPolicy = corev1.PullNever
	case "IfNotPresent":
		imagePullPolicy = corev1.PullIfNotPresent
	default:
		imagePullPolicy = corev1.PullAlways
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: deploymentAPIVersion,
			Kind:       deploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameForDeployment(&m),
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:          m.Spec.Image,
						Name:           eiContainerName,
						Ports:          exposePorts,
						LivenessProbe:  &livenessProbe,
						ReadinessProbe: &readinessProbe,
						VolumeMounts:   volMounts,

						Resources: corev1.ResourceRequirements{
							Limits:   limit,
							Requests: request,
						},

						Lifecycle: &corev1.Lifecycle{
							PreStop: getShutdownHandler(),
						},
						Env:             m.Spec.Env,
						EnvFrom:         m.Spec.EnvFrom,
						ImagePullPolicy: imagePullPolicy,
					}},
					Volumes:          volumes,
					ImagePullSecrets: imageSecrets,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   Int,
						IntVal: m.Spec.DeploySpec.MinReplicas - 1,
					},
					MaxSurge: &intstr.IntOrString{
						Type:   Int,
						IntVal: 2,
					},
				},
			},
		},
	}
	// Set Integration instance as the owner and controller
	controllerutil.SetControllerReference(&m, deployment, r.scheme)
	return deployment
}

// returns HPA for the Integration deployment with HPA version v2beta2
func createIntegrationHPA(eiConfig EIConfigNew) *v2beta2.HorizontalPodAutoscaler {

	var integration = eiConfig.integration
	owner := getOwnerDetails(eiConfig.integration)

	// target resource
	targetResource := v2beta2.CrossVersionObjectReference{
		Kind:       deploymentKind,
		Name:       nameForDeployment(&integration),
		APIVersion: deploymentAPIVersion,
	}

	// setting max replicas
	maxReplicas := integration.Spec.AutoScale.MaxReplicas

	// HPA instance for integration deployment
	hpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nameForHPA(&integration),
			Namespace:       integration.Namespace,
			OwnerReferences: owner,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MinReplicas:    &integration.Spec.DeploySpec.MinReplicas,
			MaxReplicas:    maxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        getHPAMetrics(eiConfig),
		},
	}
	return hpa
}

func getHPAMetrics(config EIConfigNew) []v2beta2.MetricSpec {
	var hpaMetricsVal = config.integrationConfigMap.Data[hpaMetricsConfigKey]
	var hpaMetrics []v2beta2.MetricSpec
	if hpaMetricsVal != "" {
		yamlErr := yaml.Unmarshal([]byte(hpaMetricsVal), &hpaMetrics)
		if yamlErr != nil {
			log.Error(yamlErr, "Error while reading HPAConfig from config")
		}
		return hpaMetrics
	}
	return nil
}

// Get livenessProbe defined in deployment file or at configmap
func getLivenessProbe(config EIConfigNew) (corev1.Probe, error) {
	if (config.integration.Spec.DeploySpec.LivenessProbe != corev1.Probe{}) {
		return config.integration.Spec.DeploySpec.LivenessProbe, nil
	} else {
		var livenessProbeVal = config.integrationConfigMap.Data[livenessProbeConfigKey]
		var livenessProbe corev1.Probe
		if livenessProbeVal != "" {
			yamlErr := yaml.Unmarshal([]byte(livenessProbeVal), &livenessProbe)
			if yamlErr != nil {
				log.Error(yamlErr, "Error while reading livenessProbe data from config")
				return corev1.Probe{}, yamlErr
			}
			return livenessProbe, nil
		}
		return corev1.Probe{}, errors.New("probe: no liveness probe defined")
	}

}

// Get readinessProbe defined in deployment file or at configmap
func getReadinessProbe(config EIConfigNew) (corev1.Probe, error) {
	if (config.integration.Spec.DeploySpec.ReadinessProbe != corev1.Probe{}) {
		return config.integration.Spec.DeploySpec.ReadinessProbe, nil
	} else {
		var readinessProbeVal = config.integrationConfigMap.Data[readinessProbeConfigKey]
		var readinessProbe corev1.Probe
		if readinessProbeVal != "" {
			yamlErr := yaml.Unmarshal([]byte(readinessProbeVal), &readinessProbe)
			if yamlErr != nil {
				log.Error(yamlErr, "Error while reading readinessProbe data from config")
				return corev1.Probe{}, yamlErr
			}
			return readinessProbe, nil
		}
		return corev1.Probe{}, errors.New("probe: no readiness probe defined in configmap")
	}
}

// Get shutdown handler specific to MI containers that handles graceful shutdown upon pod termination
func getShutdownHandler() *corev1.Handler {
	shutDownHandler := corev1.Handler{
		Exec: &corev1.ExecAction{
			Command: []string{"sh", "-c", shutdownScriptPath},
		},
	}
	return &shutDownHandler
}
