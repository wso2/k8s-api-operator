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

// Create creates the given secret in the k8s cluster
func Create(client *client.Client, secret *corev1.Secret) error {
	createErr := (*client).Create(context.TODO(), secret)
	if createErr != nil {
		logger.Error(createErr, "error creating secret", "namespace", secret.Namespace, "name", secret.Name)
	}

	return createErr
}

// Update updates the given secret in the k8s cluster
func Update(client *client.Client, secret *corev1.Secret) error {
	updateErr := (*client).Update(context.TODO(), secret)
	if updateErr != nil {
		logger.Error(updateErr, "error updating secret", "namespace", secret.Namespace, "name", secret.Name)
	}

	return updateErr
}

// Apply creates secret if not found and updates if found
func Apply(client *client.Client, secret *corev1.Secret) error {
	// get secret
	namespaceName := types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}
	err := (*client).Get(context.TODO(), namespaceName, secret)

	if err != nil && errors.IsNotFound(err) {
		return Create(client, secret)
	} else if err != nil {
		logger.Error(err, "error getting secret", "secret", namespaceName)
		return err
	}

	// secret already exists and update it
	return Update(client, secret)
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
