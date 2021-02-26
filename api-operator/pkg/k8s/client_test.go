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
	"fmt"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func getAPIObject() *wso2v1alpha2.API {

	return &wso2v1alpha2.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pet-store",
			Namespace: "api-operator",
			Labels: map[string]string{
				"label-key": "label-value",
			},
		},
	}
}

func getFakeClient(obj *wso2v1alpha2.API) *client.Client {

	objs := []runtime.Object{obj}
	s := scheme.Scheme
	s.AddKnownTypes(wso2v1alpha2.SchemeGroupVersion, obj)
	cl := fake.NewFakeClientWithScheme(s, objs...)

	return &cl
}

func TestGet(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)
	namespacedName := types.NamespacedName{Namespace: apiObject.Namespace, Name: apiObject.Name}

	err := Get(cl, namespacedName, apiObject)

	if err != nil {
		t.Error("populating the given k8s object should not return an error")
	}
}

func TestGetForNotFound(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)
	namespacedName := types.NamespacedName{Namespace: apiObject.Namespace, Name: "invalid"}

	err := Get(cl, namespacedName, apiObject)

	if err == nil {
		t.Error("populating the given k8s object for not found should return an error")
	}
}

func TestGetForInvalid(t *testing.T) {

	apiObject := &wso2v1alpha2.API{
		ObjectMeta: metav1.ObjectMeta{
		},
	}
	cl := getFakeClient(apiObject)
	namespacedName := types.NamespacedName{Namespace: "", Name: ""}

	err := Get(cl, namespacedName, apiObject)

	if err != nil {
		t.Error("populating the given k8s object for invalid should return an error")
	}
}

func TestCreate(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	newObj := &wso2v1alpha2.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pet-store-new",
			Namespace: "api-operator",
			Labels: map[string]string{
				"label-key": "label-value",
			},
		},
	}

	err := Create(cl, newObj)
	if err != nil {
		t.Error("creating a new object should not return an error")
	}
}

func TestCreateForAlreadyAvailableObject(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	err := Create(cl, apiObject)
	if err == nil {
		t.Error("creating an object that is already exist should return an error")
	}
}

func TestCreateIfNotExists(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	newObj := &wso2v1alpha2.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pet-store-new",
			Namespace: "api-operator",
			Labels: map[string]string{
				"label-key": "label-value",
			},
		},
	}

	err := CreateIfNotExists(cl, newObj)
	if err != nil {
		t.Error("creating a new object should not return an error")
	}
}

func TestCreateIfNotExistsForAlreadyAvailableObject(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	err := Create(cl, apiObject)
	if err == nil {
		t.Error("creating an object that is already exist should return an error")
	}
}

func TestApply(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	newObj := &wso2v1alpha2.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pet-store-new",
			Namespace: "api-operator",
			Labels: map[string]string{
				"label-key": "label-value",
			},
		},
	}

	err := Apply(cl, newObj)
	if err != nil {
		t.Error("applying a new object should not return an error")
	}
}

func TestApplyForAlreadyAvailableObject(t *testing.T) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	err := Apply(cl, apiObject)
	fmt.Print(err)
	if err != nil {
		t.Error("applying an object that is already exist should not return an error")
	}
}