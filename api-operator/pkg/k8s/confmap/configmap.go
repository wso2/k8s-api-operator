package confmap

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("configmap")

// Get returns a k8s Config Map with given namespacedName
func Get(client client.Client, namespacedName types.NamespacedName) (*corev1.ConfigMap, error) {
	confMap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), namespacedName, confMap)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("configmap is not found", "name", namespacedName.Name, "namespace", namespacedName.Namespace)
		return confMap, err
	} else if err != nil {
		logger.Error(err, "error getting configmap", "name", namespacedName.Name, "namespace", namespacedName.Namespace)
		return confMap, err
	}

	return confMap, nil
}
