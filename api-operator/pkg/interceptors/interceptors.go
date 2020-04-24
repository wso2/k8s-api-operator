package interceptors

import (
	"fmt"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	balIntPath  = "usr/wso2/interceptors/project-%v/"
	javaIntPath = "usr/wso2/libs/project-%v/"
)

var logger = log.Log.WithName("interceptors")

// Handle handles ballerina and java interceptors
func Handle(client *client.Client, instance *wso2v1alpha1.API) error {
	// handle ballerina interceptors
	balFound, err := handle(client, &instance.Spec.Definition.Interceptors.Ballerina, instance.Namespace, balIntPath)
	if err != nil {
		logger.Error(err, "Error handling Ballerina interceptors")
		return err
	}
	kaniko.DocFileProp.BalInterceptorsFound = balFound

	// handle java interceptors
	javaFound, err := handle(client, &instance.Spec.Definition.Interceptors.Java, instance.Namespace, javaIntPath)
	if err != nil {
		logger.Error(err, "Error handling Java interceptors")
		return err
	}
	kaniko.DocFileProp.JavaInterceptorsFound = javaFound

	return nil
}

// handle handles interceptors and returns existence of interceptors and error occurred
func handle(client *client.Client, configs *[]string, ns, mountPath string) (bool, error) {
	for i, configName := range *configs {
		// validate configmap existence
		confMap := k8s.NewConfMap()
		err := k8s.Get(client, types.NamespacedName{Namespace: ns, Name: configName}, confMap)
		if err != nil {
			logger.Error(err, "Error retrieving interceptor configmap", "configmap", confMap)
			return false, err
		}

		// mount interceptors configmap to the volume
		logger.Info("Mounting interceptor configmap to volume")
		vol, mount := k8s.ConfigMapVolumeMount(configName, fmt.Sprintf(mountPath, i))
		kaniko.AddVolume(vol, mount)
		return true, nil
	}

	return false, nil
}
