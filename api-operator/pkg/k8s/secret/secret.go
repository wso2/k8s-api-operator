package secret

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("k8s secret")

// Get returns a k8s Config Map with given namespacedName
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
