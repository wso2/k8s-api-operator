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

func NewConfMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
	}
}

func NewSecret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}
}

// NewSecretWith returns a new secret object with given namespacedName and data map
func NewSecretWith(namespacedName types.NamespacedName, data *map[string][]byte, stringData *map[string]string, owner *[]metav1.OwnerReference) *corev1.Secret {
	secret := NewSecret()
	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
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
