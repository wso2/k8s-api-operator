package deploy

import (
	"errors"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("deploy")

func SidecarContainers(client *client.Client, apiNamespace string, endpointNames *map[string]string) (*[]corev1.Container, error) {
	containerList := make([]corev1.Container, 0, len(*endpointNames))
	isAdded := make(map[string]bool)

	for endpointName, _ := range *endpointNames {
		// deploy sidecar only if endpoint name is not empty and not already deployed
		if endpointName != "" && !isAdded[endpointName] {
			targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
			erCr := k8s.Get(client,
				types.NamespacedName{Namespace: apiNamespace, Name: endpointName}, targetEndpointCr)
			if erCr == nil && targetEndpointCr.Spec.Deploy.DockerImage != "" {
				sidecarContainer := corev1.Container{
					Image: targetEndpointCr.Spec.Deploy.DockerImage,
					Name:  targetEndpointCr.Spec.Deploy.Name,
					Ports: []corev1.ContainerPort{{
						ContainerPort: targetEndpointCr.Spec.Port,
					}},
				}
				logger.Info("Added sidecar container to the list of containers to be deployed",
					"endpoint-name", endpointName, "docker-image", targetEndpointCr.Spec.Deploy.DockerImage)
				containerList = append(containerList, sidecarContainer)
				isAdded[endpointName] = true
			} else {
				err := erCr
				if erCr == nil {
					err = errors.New("docker image of the endpoint is empty")
				}

				logger.Error(err, "Failed to deploy the sidecar endpoint", "endpoint name", endpointName)
				return nil, err
			}
		}
	}

	return &containerList, nil
}
