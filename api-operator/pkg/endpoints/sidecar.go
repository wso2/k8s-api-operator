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
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("endpoints.sidecar")

func AddSidecarContainers(client *client.Client, apiNamespace string, endpointNames *map[string]string) error {
	containerList := make([]corev1.Container, 0, len(*endpointNames))
	isAdded := make(map[string]bool)

	for endpointName := range *endpointNames {
		// deploy sidecar only if endpoint name is not empty and not already deployed
		if endpointName != "" && !isAdded[endpointName] {
			targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
			erCr := k8s.Get(client,
				types.NamespacedName{Namespace: apiNamespace, Name: endpointName}, targetEndpointCr)
			if erCr == nil && targetEndpointCr.Spec.Deploy.DockerImage != "" {
				// set container ports
				containerPorts := make([]corev1.ContainerPort, 0, len(targetEndpointCr.Spec.Ports))
				for _, port := range targetEndpointCr.Spec.Ports {
					containerPorts = append(containerPorts, corev1.ContainerPort{
						Name:          port.Name,
						ContainerPort: port.TargetPort,
					})
				}

				sidecarContainer := corev1.Container{
					Image: targetEndpointCr.Spec.Deploy.DockerImage,
					Name:  targetEndpointCr.Spec.Deploy.Name,
					Ports: containerPorts,
				}
				logger.Info("Added sidecar container to the list of containers to be deployed",
					"endpoint_name", endpointName, "docker_image", targetEndpointCr.Spec.Deploy.DockerImage)
				containerList = append(containerList, sidecarContainer)
				isAdded[endpointName] = true
			} else {
				err := erCr
				if erCr == nil {
					err = errors.New("docker image of the endpoint is empty")
				}

				logger.Error(err, "Failed to deploy the sidecar endpoint", "endpoint_name", endpointName)
				return err
			}
		}
	}

	mgw.AddContainers(&containerList)
	return nil
}
