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

package k8s

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("k8s configmap")

// Get populates the given k8s object with k8s cluster object values in the given namespacedName
func Get(client *client.Client, namespacedName types.NamespacedName, obj runtime.Object) error {
	err := (*client).Get(context.TODO(), namespacedName, obj)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("k8s object is not found", "object", obj)
		return err
	} else if err != nil {
		logger.Error(err, "error getting k8s object", "object", obj)
		return err
	}

	logger.Info("getting k8s object is success", "object", obj)
	return nil
}

// Create creates the given k8s object in the k8s cluster
func Create(client *client.Client, obj runtime.Object) error {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	err := (*client).Create(context.TODO(), obj)
	if err != nil {
		logger.Error(err, "error creating k8s object", "object", obj)
	}
	logger.Info("creating k8s object is success", "kind", kind, "object", obj)
	return err
}

// Update updates the given k8s object in the k8s cluster
func Update(client *client.Client, obj runtime.Object) error {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	updateErr := (*client).Update(context.TODO(), obj)
	if updateErr != nil {
		logger.Error(updateErr, "error updating configmap", "object", obj)
	}
	logger.Info("updating k8s object is success", "kind", kind, "object", obj)
	return updateErr
}

// Apply creates k8s object if not found and updates if found
func Apply(client *client.Client, obj runtime.Object) error {
	// get k8s object
	objMeta := obj.(metav1.Object)
	namespaceName := types.NamespacedName{Namespace: objMeta.GetNamespace(), Name: objMeta.GetName()}
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	err := (*client).Get(context.TODO(), namespaceName, obj)

	if err != nil && errors.IsNotFound(err) {
		return Create(client, obj)
	} else if err != nil {
		logger.Error(err, "error applying k8s object while getting it from cluster", "kind", kind, "object", obj)
		return err
	}

	// configmap already exists and update it
	return Update(client, obj)
}

// UpdateOwner updates the k8s object with the owner reference
func UpdateOwner(client *client.Client, owner *[]metav1.OwnerReference, obj runtime.Object) error {
	obj.(metav1.Object).SetOwnerReferences(*owner)
	kind := obj.GetObjectKind().GroupVersionKind().Kind

	err := (*client).Update(context.TODO(), obj)
	if err != nil {
		logger.Error(err, "error updating owner reference of k8s object", "kind", kind, "object", obj)
	}
	return err
}
