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

package api

import (
	"context"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy"
	"strconv"
	"time"

	"github.com/wso2/k8s-api-operator/api-operator/pkg/apim"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("api.controller")

// Add creates a new API Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAPI{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("api-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		"api-controller",
		mgr,
		controller.Options{
			MaxConcurrentReconciles: 10,
			Reconciler:              r,
		},
	)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource API
	err = c.Watch(&source.Kind{Type: &wso2v1alpha2.API{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner API
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha2.API{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAPI{}

// ReconcileAPI reconciles a API object
type ReconcileAPI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a API object and makes changes based on the state read
// and what is in the API.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("request_namespace", request.Namespace, "request_name", request.Name)
	reqLogger.Info("Reconciling API")

	instance := &wso2v1alpha2.API{}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	// Request info
	requestInfo := &common.RequestInfo{Request: request, Client: r.client, Object: instance, Log: log, EvnRecorder:
	r.recorder}
	ctx = requestInfo.NewContext(ctx)

	err := k8s.Get(&r.client, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//get configurations file for the controller
	controlConf := k8s.NewConfMap()
	errConf := k8s.Get(&r.client, types.NamespacedName{Namespace: config.SystemNamespace, Name: controllerConfName},
		controlConf)

	if errConf != nil {
		if errors.IsNotFound(errConf) {
			// Required configmap is not found. User should add the required config to proceed.
			// Return and requeue
			reqLogger.Error(errConf, "Required configmap is not found. Requeue request after 10 seconds")
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, errConf
	}

	controlConfigData := controlConf.Data
	instance.Status.Replicas = instance.Spec.Replicas
	importAPIEnabled, err := strconv.ParseBool(controlConfigData[importAPIEnabledConst])
	if err != nil {
		reqLogger.Error(err, "Invalid boolean value for importAPIEnabled")
		return reconcile.Result{}, err
	}
	if importAPIEnabled {
		importErr := apim.ImportAPI(&r.client, instance)
		if importErr != nil {
			r.recorder.Event(instance, eventTypeError, "FailedAPIImport",
				fmt.Sprintf("Error occured while importing the API to APIM"))
			return reconcile.Result{}, importErr
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, "APIImport",
			fmt.Sprintf("Successfully imported the API to APIM"))
		reqLogger.Info("Successfully imported the API to APIM", "api_name", instance.Name)
	}

	// Deploy the API to MGW Adapter
	deployErr := envoy.DeployAPItoMgw(&r.client, instance)
	if deployErr != nil {
		r.recorder.Event(instance, eventTypeError, "FailedAPIDeployToMGW",
			fmt.Sprintf("Error occured while deploying API to Envoy MGW Adapter"))
		return reconcile.Result{}, deployErr
	}
	r.recorder.Event(instance, corev1.EventTypeNormal, "APIDeploy",
		fmt.Sprintf("Successfully deployed API to Envoy MGW Adapter"))
	reqLogger.Info("Successfully deployed API to Envoy MGW Adapter", "api_name", instance.Name)

	apiList := &wso2v1alpha2.APIList{}
	if err := requestInfo.Client.List(ctx, apiList, client.InNamespace(common.WatchNamespace)); err != nil {
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Error reading all ingresses in the specified namespace", "namespace",
			common.WatchNamespace)
		return reconcile.Result{}, err
	}
	err = envoy.CreateFileToSend(apiList, &r.client)
	if err != nil {
		reqLogger.Error(err, "Error 67!!!")
	}

	return reconcile.Result{}, nil
}
