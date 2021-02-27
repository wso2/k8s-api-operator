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
	"errors"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
	"time"
)

const finalizer = "wso2.microgateway/api.finalizer"

func getContext() context.Context {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	return ctx
}

func getRequestInfo() (*common.RequestInfo, *wso2v1alpha2.API) {

	apiObject := getAPIObject()
	cl := getFakeClient(apiObject)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      apiObject.Name,
			Namespace: apiObject.Namespace,
		},
	}

	var log = logf.Log.WithName("api.controller")
	reqLogger := log.WithValues("request_namespace", req.Namespace, "request_name", req.Name)

	requestInfo := &common.RequestInfo{Request: req, Client: *cl, Object: apiObject, Log: reqLogger,
		EvnRecorder: nil}

	return requestInfo, apiObject
}

func TestHandleDeletion(t *testing.T) {

	requestInfo, apiObject := getRequestInfo()
	_, _, err := HandleDeletion(apiObject, getContext(), requestInfo, finalizer, nil)

	if err != nil {
		t.Error("delete finalizer for available object should not return an error")
	}
}

func TestHandleDeletionWithFinalizer(t *testing.T) {

	requestInfo, apiObject := getRequestInfo()
	finalizers := []string{finalizer}
	apiObject.SetFinalizers(finalizers)
	_, _, err := HandleDeletion(apiObject, getContext(), requestInfo, finalizer, nil)

	if err != nil {
		t.Error("delete finalizer for available object (Same Finalizer) should not return an error")
	}
}

func TestHandleDeletionForDeletedObj(t *testing.T) {

	requestInfo, apiObject := getRequestInfo()
	deletionTime := metav1.Date(2020, time.January, 26, 15, 45, 40, 00, time.UTC)
	apiObject.ObjectMeta.SetDeletionTimestamp(&deletionTime)

	_, _, err := HandleDeletion(apiObject, getContext(), requestInfo, finalizer, nil)

	if err != nil {
		t.Error("delete finalizer for already deleted object should not return an error")
	}
}

func TestHandleDeletionForDeletedObjWithFinalizer(t *testing.T) {

	requestInfo, apiObject := getRequestInfo()
	deletionTime := metav1.Date(2020, time.January, 26, 15, 45, 40, 00, time.UTC)
	apiObject.ObjectMeta.SetDeletionTimestamp(&deletionTime)
	finalizers := []string{finalizer}
	apiObject.SetFinalizers(finalizers)

	_, _, err := HandleDeletion(apiObject, getContext(), requestInfo, finalizer, handleAPI)

	if err != nil {
		t.Error("delete finalizer for already deleted object for same finalizer should not return an error")
	}
}

func TestHandleDeletionForDeletedObjWithFinalizerErrorCase(t *testing.T) {

	requestInfo, apiObject := getRequestInfo()
	deletionTime := metav1.Date(2020, time.January, 26, 15, 45, 40, 00, time.UTC)
	apiObject.ObjectMeta.SetDeletionTimestamp(&deletionTime)
	finalizers := []string{finalizer}
	apiObject.SetFinalizers(finalizers)

	_, _, err := HandleDeletion(apiObject, getContext(), requestInfo, finalizer, handleAPIWithError)

	if err == nil {
		t.Error("delete finalizer for already deleted object for same finalizer with error should " +
			"return an error")
	}
}

func handleAPI(*wso2v1alpha2.API) error {
	return nil
}

func handleAPIWithError(*wso2v1alpha2.API) error {
	return errors.New("error while handling API")
}
