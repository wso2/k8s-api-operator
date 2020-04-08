package confmap

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("configmap")

// Get returns a k8s Config Map with given namespacedName
func Get(client *client.Client, namespacedName types.NamespacedName) (*corev1.ConfigMap, error) {
	confMap := &corev1.ConfigMap{}
	err := (*client).Get(context.TODO(), namespacedName, confMap)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("configmap is not found", "namespace", namespacedName.Namespace, "name", namespacedName.Name)
		return confMap, err
	} else if err != nil {
		logger.Error(err, "error getting configmap", "namespace", namespacedName.Namespace, "name", namespacedName.Name)
		return confMap, err
	}

	return confMap, nil
}

// UpdateOwner updates the config map with the owner reference
func UpdateOwner(client *client.Client, owner []metav1.OwnerReference, configMap *corev1.ConfigMap) error {
	configMap.OwnerReferences = owner

	err := (*client).Update(context.TODO(), configMap)
	if err != nil {
		logger.Error(err, "error updating owner reference for configmap", "namespace", configMap.Namespace, "name", configMap.Name)
	}
	return err
}

// New returns a new configmap object with given namespacedName and data map
func New(namespacedName types.NamespacedName, dataMap *map[string]string, owner []metav1.OwnerReference) *corev1.ConfigMap {

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namespacedName.Name,
			Namespace:       namespacedName.Namespace,
			OwnerReferences: owner,
		},
		Data: *dataMap,
	}
}
