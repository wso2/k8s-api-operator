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

var logger = log.Log.WithName("k8s configmap")

// Get returns a k8s Config Map with given namespacedName
func Get(client *client.Client, namespacedName types.NamespacedName) (*corev1.ConfigMap, error) {
	confMap := &corev1.ConfigMap{}
	err := (*client).Get(context.TODO(), namespacedName, confMap)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("configmap is not found", "configmap", namespacedName)
		return confMap, err
	} else if err != nil {
		logger.Error(err, "error getting configmap", "configmap", namespacedName)
		return confMap, err
	}

	return confMap, nil
}

// Create creates the given configmap in the k8s cluster
func Create(client *client.Client, confMap *corev1.ConfigMap) error {
	createErr := (*client).Create(context.TODO(), confMap)
	if createErr != nil {
		logger.Error(createErr, "error creating configmap", "namespace", confMap.Namespace, "name", confMap.Name)
	}

	return createErr
}

// Update updates the given configmap in the k8s cluster
func Update(client *client.Client, confMap *corev1.ConfigMap) error {
	updateErr := (*client).Update(context.TODO(), confMap)
	if updateErr != nil {
		logger.Error(updateErr, "error updating configmap", "namespace", confMap.Namespace, "name", confMap.Name)
	}

	return updateErr
}

// Apply creates configmap if not found and updates if found
func Apply(client *client.Client, confMap *corev1.ConfigMap) error {
	// get configmap
	namespaceName := types.NamespacedName{Namespace: confMap.Namespace, Name: confMap.Name}
	err := (*client).Get(context.TODO(), namespaceName, confMap)

	if err != nil && errors.IsNotFound(err) {
		return Create(client, confMap)
	} else if err != nil {
		logger.Error(err, "error getting configmap", "configmap", namespaceName)
		return err
	}

	// configmap already exists and update it
	return Update(client, confMap)
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

// Copy copies config map from given namespacedName to destination namespacedName
func Copy(client *client.Client, fromNsName, toNsName types.NamespacedName) error {
	// Get configmap
	fromCnf, fromErr := Get(client, fromNsName)
	if fromErr != nil {
		logger.Error(fromErr, "error coping configmap", "configmap", fromNsName)
		return fromErr
	}

	toCnf := &corev1.ConfigMap{}
	toCnf.Data = fromCnf.Data
	toCnf.BinaryData = fromCnf.BinaryData
	toCnf.Namespace = toNsName.Namespace
	toCnf.Name = toNsName.Name

	return Apply(client, toCnf)
}
