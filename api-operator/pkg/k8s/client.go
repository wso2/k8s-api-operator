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

var logCnt = log.Log.WithName("k8s.client")

// Get populates the given k8s object with k8s cluster object values in the given namespacedName
func Get(client *client.Client, namespacedName types.NamespacedName, obj runtime.Object) error {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	err := (*client).Get(context.TODO(), namespacedName, obj)

	if err != nil && errors.IsNotFound(err) {
		logCnt.Info("K8s object is not found", "kind", kind, "key", namespacedName)
		return err
	} else if err != nil {
		logCnt.Error(err, "Error getting k8s object", "kind", kind, "key", namespacedName)
		return err
	}
	logCnt.Info("Getting k8s object is success", "object", obj)
	return nil
}

// Create creates the given k8s object in the k8s cluster
func Create(client *client.Client, obj runtime.Object) error {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	err := (*client).Create(context.TODO(), obj)
	if err != nil {
		logCnt.Error(err, "Error creating k8s object", "object", obj)
	}
	logCnt.Info("Creating k8s object is success", "kind", kind, "object", obj)
	return err
}

// CreateIfNotExists creates the given k8s object if the object is not exists in the k8s cluster
func CreateIfNotExists(client *client.Client, obj runtime.Object) error {
	// get k8s object
	objMeta := obj.(metav1.Object)
	namespaceName := types.NamespacedName{Namespace: objMeta.GetNamespace(), Name: objMeta.GetName()}
	kind := obj.GetObjectKind().GroupVersionKind().Kind

	err := (*client).Get(context.TODO(), namespaceName, obj)
	if err != nil && errors.IsNotFound(err) {
		return Create(client, obj)
	} else if err != nil {
		logCnt.Error(err, "Error creating k8s object if not found. Error while getting object", "kind", kind, "object", obj)
	}
	return err
}

// Apply creates k8s object if not found and updates if found
func Apply(client *client.Client, obj runtime.Object) error {
	// get k8s object
	kind := obj.GetObjectKind().GroupVersionKind().Kind

	err := (*client).Update(context.TODO(), obj)
	if err != nil && errors.IsNotFound(err) {
		return Create(client, obj)
	} else if err != nil {
		logCnt.Error(err, "Error applying k8s object while getting it from cluster", "kind", kind, "object", obj)
	}
	return err
}

