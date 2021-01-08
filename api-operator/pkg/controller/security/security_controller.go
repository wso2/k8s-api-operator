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
package security

import (
	"context"
	"strings"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_security")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Security Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSecurity{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("security-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Security
	err = c.Watch(&source.Kind{Type: &wso2v1alpha1.Security{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Security
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.Security{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSecurity{}

// ReconcileSecurity reconciles a Security object
type ReconcileSecurity struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Security object and makes changes based on the state read
// and what is in the Security.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSecurity) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Security")

	// Fetch the Security instance
	instance := &wso2v1alpha1.Security{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	userNamespace := instance.Namespace
	if strings.EqualFold(instance.Spec.Type, "JWT") {
		for _, securityConfig := range instance.Spec.SecurityConfig {
			if securityConfig.Issuer == "" {
				reqLogger.Error(err, "Required fields are missing")
				return reconcile.Result{}, err
			}
			if securityConfig.Audience == "" {
				reqLogger.Info("Audience is not Provided")
			}

			if securityConfig.JwksURL == "" {
				certificateSecret := &corev1.Secret{}
				errcertificate := r.client.Get(context.TODO(), types.NamespacedName{Name: securityConfig.Certificate, Namespace: userNamespace}, certificateSecret)

				if errcertificate != nil && errors.IsNotFound(errcertificate) {
					reqLogger.Info("defined secret for cretificate is not found")
					return reconcile.Result{}, errcertificate
				}
			}
		}
	}

	if strings.EqualFold(instance.Spec.Type, "apiKey") {
		for _, securityConfig := range instance.Spec.SecurityConfig {
			if securityConfig.Issuer == "" || securityConfig.Audience == "" || securityConfig.Alias == "" {
				reqLogger.Error(err, "Required fields are missing")
				return reconcile.Result{}, err
			}
		}
	}

	if strings.EqualFold(instance.Spec.Type, "Oauth") {
		for _, securityConfig := range instance.Spec.SecurityConfig {
			if securityConfig.Credentials == "" || securityConfig.Endpoint == "" {
				reqLogger.Error(err, "Required fields are missing")
				return reconcile.Result{}, err
			}

			if securityConfig.JwksURL == "" {
				credentialSecret := &corev1.Secret{}
				errcertificate := r.client.Get(context.TODO(), types.NamespacedName{Name: securityConfig.Credentials, Namespace: userNamespace}, credentialSecret)

				if errcertificate != nil && errors.IsNotFound(errcertificate) {
					reqLogger.Info("defined secret for credentials is not found")
					return reconcile.Result{}, errcertificate
				}
			}
		}
	}

	if strings.EqualFold(instance.Spec.Type, "Basic") {
		for _, securityConfig := range instance.Spec.SecurityConfig {
			if securityConfig.Credentials == "" {
				reqLogger.Error(err, "Required fields are missing")
				return reconcile.Result{}, err
			}
		}
	}
	return reconcile.Result{}, nil
}
