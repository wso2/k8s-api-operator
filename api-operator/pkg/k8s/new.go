package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetNewConfMap returns a new configmap object with given namespacedName and data map
func GetNewConfMap(namespacedName types.NamespacedName, dataMap *map[string]string, binaryData *map[string][]byte, owner *[]metav1.OwnerReference) *corev1.ConfigMap {
	confMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
	}

	if owner != nil {
		confMap.OwnerReferences = *owner
	}
	if dataMap != nil {
		confMap.Data = *dataMap
	}
	if binaryData != nil {
		confMap.BinaryData = *binaryData
	}

	return confMap
}

// GetNewSecret returns a new secret object with given namespacedName and data map
func GetNewSecret(namespacedName types.NamespacedName, data *map[string][]byte, stringData *map[string]string, owner *[]metav1.OwnerReference) *corev1.Secret {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}

	if owner != nil {
		secret.OwnerReferences = *owner
	}
	if data != nil {
		secret.Data = *data
	}
	if stringData != nil {
		secret.StringData = *stringData
	}

	return secret
}

// GetNewOwnerRef returns an array with a new owner reference object of given meta data
func GetNewOwnerRef(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) *[]metav1.OwnerReference {
	setOwner := true
	return &[]metav1.OwnerReference{
		{
			APIVersion:         typeMeta.APIVersion,
			Kind:               typeMeta.Kind,
			Name:               objectMeta.Name,
			UID:                objectMeta.UID,
			Controller:         &setOwner,
			BlockOwnerDeletion: &setOwner,
		},
	}
}
