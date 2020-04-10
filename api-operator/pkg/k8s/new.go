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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewConfMap returns a new configmap object with given namespacedName and data map
func NewConfMap(namespacedName types.NamespacedName, dataMap *map[string]string, binaryData *map[string][]byte, owner *[]metav1.OwnerReference) *corev1.ConfigMap {
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

// NewSecret returns a new secret object with given namespacedName and data map
func NewSecret(namespacedName types.NamespacedName, data *map[string][]byte, stringData *map[string]string, owner *[]metav1.OwnerReference) *corev1.Secret {
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

// NewOwnerRef returns an array with a new owner reference object of given meta data
func NewOwnerRef(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) *[]metav1.OwnerReference {
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
