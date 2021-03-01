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
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"testing"
)

func TestNewConfMap(t *testing.T) {

	var config *coreV1.ConfigMap

	config = NewConfMap()
	if config == nil {
		t.Error("generating a new configmap should not return a nil config")
	}
}

func TestNewSecret(t *testing.T) {

	var secret *coreV1.Secret

	secret = NewSecret()
	if secret == nil {
		t.Error("generating a new secret should not return a nil secret")
	}
}

func TestNewSecretWith(t *testing.T) {

	dataValues := []byte("value1")
	var data = make(map[string][]byte)
	data["key1"] = dataValues

	var stringData = make(map[string]string)
	stringData["key2"] = "value2"

	owner := metav1.OwnerReference{
		APIVersion:         "v1",
		Kind:               "API",
		Name:               "test",
		UID:                "uuid",
		BlockOwnerDeletion: pointer.BoolPtr(true),
		Controller:         pointer.BoolPtr(true),
	}
	ownerRefs := make([]metav1.OwnerReference, 0, 1)
	ownerRefs = append(ownerRefs, owner)

	namespacedName := types.NamespacedName{Namespace: "test-ns", Name: "test-secret"}

	secret := NewSecretWith(namespacedName, &data, &stringData, &ownerRefs)
	if secret == nil {
		t.Error("generating a new secret with all values should not return a nil secret")
	}
}

func TestNewSecretWithOwnerNil(t *testing.T) {

	dataValues := []byte("value1")
	var data = make(map[string][]byte)
	data["key1"] = dataValues

	var stringData = make(map[string]string)
	stringData["key2"] = "value2"

	namespacedName := types.NamespacedName{Namespace: "test-ns", Name: "test-secret"}

	secret := NewSecretWith(namespacedName, &data, &stringData, nil)
	if secret == nil {
		t.Error("generating a new secret with owner nil should not return a nil secret")
	}

	if secret.OwnerReferences != nil {
		t.Error("generating a new secret with owner nil should return a nil owner reference")
	}
}

func TestNewSecretWithDataNil(t *testing.T) {

	var stringData = make(map[string]string)
	stringData["key2"] = "value2"

	namespacedName := types.NamespacedName{Namespace: "test-ns", Name: "test-secret"}

	secret := NewSecretWith(namespacedName, nil, &stringData, nil)
	if secret == nil {
		t.Error("generating a new secret with data nil should not return a nil secret")
	}

	if secret.Data != nil {
		t.Error("generating a new secret with data nil should return a nil data")
	}
}

func TestNewSecretWithStringDataNil(t *testing.T) {

	namespacedName := types.NamespacedName{Namespace: "test-ns", Name: "test-secret"}

	secret := NewSecretWith(namespacedName, nil, nil, nil)
	if secret == nil {
		t.Error("generating a new secret with string data nil should not return a nil secret")
	}

	if secret.StringData != nil {
		t.Error("generating a new secret with string data nil should return a nil string data")
	}
}

func TestNewSecretWithAllDataNil(t *testing.T) {

	namespacedName := types.NamespacedName{Namespace: "test-ns", Name: "test-secret"}

	owner := metav1.OwnerReference{
		APIVersion:         "v1",
		Kind:               "API",
		Name:               "test",
		UID:                "uuid",
		BlockOwnerDeletion: pointer.BoolPtr(true),
		Controller:         pointer.BoolPtr(true),
	}
	ownerRefs := make([]metav1.OwnerReference, 0, 1)
	ownerRefs = append(ownerRefs, owner)

	secret := NewSecretWith(namespacedName, nil, nil, &ownerRefs)
	if secret == nil {
		t.Error("generating a new secret with all data nil should not return a nil secret")
	}

	if secret.OwnerReferences == nil {
		t.Error("generating a new secret with all data nil should not return a nil owner reference")
	}
}