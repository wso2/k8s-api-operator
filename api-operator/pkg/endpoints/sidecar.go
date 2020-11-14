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

package endpoints

import (
	"errors"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	resourceRequestCPUTarget    = "resourceRequestCPUTarget"
	resourceRequestMemoryTarget = "resourceRequestMemoryTarget"
	resourceLimitCPUTarget      = "resourceLimitCPUTarget"
	resourceLimitMemoryTarget   = "resourceLimitMemoryTarget"
)

var logger = log.Log.WithName("endpoints.sidecar")

func GetSidecarContainers(client *client.Client, apiNamespace string, sidecarEpNames *map[string]bool) ([]corev1.Container, error) {
	containerList := make([]corev1.Container, 0, len(*sidecarEpNames))
	isAdded := make(map[string]bool)

	for sidecarEpName := range *sidecarEpNames {
		// deploy sidecar only if endpoint name is not empty and not already deployed
		if sidecarEpName != "" && !isAdded[sidecarEpName] {
			targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
			erCr := k8s.Get(client,
				types.NamespacedName{Namespace: apiNamespace, Name: sidecarEpName}, targetEndpointCr)
			if erCr == nil && targetEndpointCr.Spec.Deploy.DockerImage != "" {
				// set container ports
				containerPorts := make([]corev1.ContainerPort, 0, len(targetEndpointCr.Spec.Ports))
				for _, port := range targetEndpointCr.Spec.Ports {
					containerPorts = append(containerPorts, corev1.ContainerPort{
						Name:          port.Name,
						ContainerPort: port.TargetPort,
					})
				}

				resourceRequirements, resourceLimits, err := getResourceMetadata(client, targetEndpointCr)
				if err != nil {
					if k8sError.IsNotFound(err) {
						// Controller configmap is not found.
						logger.Error(err, "Controller configuration file is not found")
						return nil, err
					}
					// Error reading the object
					return nil, err
				}

				sidecarContainer := corev1.Container{
					Image: targetEndpointCr.Spec.Deploy.DockerImage,
					Name:  targetEndpointCr.Spec.Deploy.Name,
					Ports: containerPorts,
					Resources: corev1.ResourceRequirements{
						Limits:   resourceLimits,
						Requests: resourceRequirements,
					},
				}
				logger.Info("Added sidecar container to the list of containers to be deployed",
					"endpoint_name", sidecarEpName, "docker_image", targetEndpointCr.Spec.Deploy.DockerImage)
				containerList = append(containerList, sidecarContainer)
				isAdded[sidecarEpName] = true
			} else {
				err := erCr
				if erCr == nil {
					err = errors.New("docker image of the endpoint is empty")
				}

				logger.Error(err, "Failed to deploy the sidecar endpoint", "endpoint_name", sidecarEpName)
				return nil, err
			}
		}
	}

	return containerList, nil
}

func getResourceMetadata(client *client.Client,
	targetEndpointCr *wso2v1alpha1.TargetEndpoint) (corev1.ResourceList, corev1.ResourceList, error) {
	controllerConfMap := &corev1.ConfigMap{}
	err := k8s.Get(client,
		types.NamespacedName{Namespace: config.SystemNamespace, Name: "controller-config"}, controllerConfMap)
	if err != nil {
		return nil, nil, err
	}
	controlConfigData := controllerConfMap.Data
	getResourceReqCPU := controlConfigData[resourceRequestCPUTarget]
	getResourceReqMemory := controlConfigData[resourceRequestMemoryTarget]
	getResourceLimitCPU := controlConfigData[resourceLimitCPUTarget]
	getResourceLimitMemory := controlConfigData[resourceLimitMemoryTarget]

	var reqCpu, reqMemory, limitCpu, limitMemory string

	if targetEndpointCr.Spec.Deploy.ReqCpu != "" {
		reqCpu = targetEndpointCr.Spec.Deploy.ReqCpu
	} else {
		reqCpu = getResourceReqCPU
	}

	if targetEndpointCr.Spec.Deploy.ReqMemory != "" {
		reqMemory = targetEndpointCr.Spec.Deploy.ReqMemory
	} else {
		reqMemory = getResourceReqMemory
	}

	if targetEndpointCr.Spec.Deploy.LimitCpu != "" {
		limitCpu = targetEndpointCr.Spec.Deploy.LimitCpu
	} else {
		limitCpu = getResourceLimitCPU
	}

	if targetEndpointCr.Spec.Deploy.MemoryLimit != "" {
		limitMemory = targetEndpointCr.Spec.Deploy.MemoryLimit
	} else {
		limitMemory = getResourceLimitMemory
	}

	resourceRequirements := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(reqCpu),
		corev1.ResourceMemory: resource.MustParse(reqMemory),
	}
	resourceLimits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(limitCpu),
		corev1.ResourceMemory: resource.MustParse(limitMemory),
	}

	return resourceRequirements, resourceLimits, nil
}
