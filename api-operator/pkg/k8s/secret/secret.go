package secret

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("k8s secret")

// Get returns a k8s secret with given namespacedName
func Get(client *client.Client, namespacedName types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := (*client).Get(context.TODO(), namespacedName, secret)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("secret is not found", "secret", namespacedName)
		return secret, err
	} else if err != nil {
		logger.Error(err, "error getting configmap", "secret", namespacedName)
		return secret, err
	}

	return secret, nil
}

// New returns a new secret object with given namespacedName and data map
func New(namespacedName types.NamespacedName, data *map[string][]byte, stringData *map[string]string, owner []metav1.OwnerReference) *corev1.Secret {

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namespacedName.Name,
			Namespace:       namespacedName.Namespace,
			OwnerReferences: owner,
		},
		Data:       *data,
		StringData: *stringData,
	}
}
