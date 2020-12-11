/*
 * Copyright (c) 2020 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// deploymentForIntegration returns a integration Deployment object
func (r *ReconcileIntegration) deploymentForIntegration(m *wso2v1alpha1.Integration) *appsv1.Deployment {
	//set HTTP and HTTPS ports for as container ports
	exposePorts := []corev1.ContainerPort{
		corev1.ContainerPort{
			ContainerPort: 8290,
		},
	}

	// check inbound endpoint port is exist and append to the container port
	for _, port := range m.Spec.InboundPorts {
		exposePorts = append(
			exposePorts,
			corev1.ContainerPort{
				ContainerPort: port,
			},
		)
	}

	// check ImagePullPolicy has given with the integration
	imageSecrets := []corev1.LocalObjectReference{}
	if m.Spec.ImagePullSecret != "" {
		imageSecrets = append(
			imageSecrets,
			corev1.LocalObjectReference{
				Name: m.Spec.ImagePullSecret,
			},
		)
	}

	labels := labelsForIntegration(m.Name)
	replicas := m.Spec.Replicas

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameForDeployment(m),
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
						Image:           m.Spec.Image,
						Name:            "micro-integrator",
						Ports:           exposePorts,
						Env:             m.Spec.Env,
						EnvFrom: 	 m.Spec.EnvFrom,
						ImagePullPolicy: corev1.PullAlways,
					}},
					ImagePullSecrets: imageSecrets,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   Int,
						IntVal: m.Spec.Replicas - 1,
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
	controllerutil.SetControllerReference(m, deployment, r.scheme)
	return deployment
}
