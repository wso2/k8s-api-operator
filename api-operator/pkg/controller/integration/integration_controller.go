/*
 * Copyright (c) 2020 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package integration

import (
	"context"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
)

var log = logf.Log.WithName("controller_integration")

//define type const
const (
	//define type Int and String
	Int intstr.Type = iota
	String
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Integration Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIntegration{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("integration-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Integration
	err = c.Watch(&source.Kind{Type: &wso2v1alpha2.Integration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Integration
	// Watch for deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha2.Integration{},
	})

	// Watch for service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha2.Integration{},
	})

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIntegration implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIntegration{}

// ReconcileIntegration reconciles a Integration object
type ReconcileIntegration struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Integration object and makes changes based on the state read
// and what is in the Integration.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIntegration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Integration")

	// Fetch the Integration integration
	integration := &wso2v1alpha2.Integration{}
	err := r.client.Get(context.TODO(), request.NamespacedName, integration)
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

	// Create ei config struct using default ei configmap yaml
	eiConfig := r.UpdateDefaultConfigs(integration)

	// Check if the deployment already exists, if not create a new one
	deploymentObj := &appsv1.Deployment{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: nameForDeployment(integration), Namespace: integration.Namespace}, deploymentObj)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		deployment := r.deploymentForIntegration(integration, eiConfig)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the deployment replicas is the same as the spec
	replicas := integration.Spec.DeploySpec.MinReplicas
	if *deploymentObj.Spec.Replicas != replicas {
		deploymentObj.Spec.Replicas = &replicas
		err = r.client.Update(context.TODO(), deploymentObj)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", deploymentObj.Namespace, "Deployment.Name", deploymentObj.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Check auto scaling and create/update horizontal autoscaler for the deployment
	var autoScaleEnabled bool
	if integration.Spec.AutoScale.Enabled != "" {
		autoScaleEnabled, _ = strconv.ParseBool(integration.Spec.AutoScale.Enabled)
	} else {
		autoScaleEnabled = eiConfig.EnableAutoScale
	}
	if autoScaleEnabled {
		owner := getOwnerDetails(integration)
		hpa := createIntegrationHPA(integration, deploymentObj, eiConfig, owner)
		// create or apply HPA
		err = k8s.Apply(&r.client, hpa)
		if err != nil {
			reqLogger.Info("Failed to create/update HPA", "HPA.Namespace", hpa.Namespace, "HPA.Name", hpa.Name)
			return reconcile.Result{}, err
		}
	}

	// Check if the service already exists, if not create a new one
	serviceObj := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: nameForService(integration), Namespace: integration.Namespace}, serviceObj)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		service := r.serviceForIntegration(integration)
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	// Check if the ingress already exists, if not create a new one, if yes update it
	if eiConfig.AutoCreateIngress != false {
		ingress := &v1beta1.Ingress{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: nameForIngress(), Namespace: integration.Namespace}, ingress)
		if err != nil && errors.IsNotFound(err) {
			// Define a new Ingress
			eiIngress := r.ingressForIntegration(integration, &eiConfig)
			reqLogger.Info("Creating a new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
			err = r.client.Create(context.TODO(), eiIngress)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
				return reconcile.Result{}, err
			}
			// Ingress created successfully - return and requeue
			reqLogger.Info("Ingress created successfully")

		} else if err == nil {
			_, ruleExists := CheckIngressRulesExist(integration, &eiConfig, ingress)
			if !ruleExists {
				eiIngress := r.updateIngressForIntegration(integration, &eiConfig, ingress)
				reqLogger.Info("Updating a new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
				err = r.client.Update(context.TODO(), eiIngress)
				if err != nil {
					reqLogger.Error(err, "Failed to updated new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
					return reconcile.Result{}, err
				}
				// Ingress updated successfully - return and requeue
				reqLogger.Info("Ingress updated successfully")
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Ingress")
			return reconcile.Result{}, err
		}
	}

	// Update status.Status if needed
	availableReplicas := deploymentObj.Status.AvailableReplicas
	currentStatus := "NotRunning"
	if availableReplicas > 0 {
		currentStatus = "Running"
	}
	if !reflect.DeepEqual(currentStatus, integration.Status.Readiness) {
		integration.Status.Readiness = currentStatus
		err := r.client.Status().Update(context.TODO(), integration)
		if err != nil {
			reqLogger.Error(err, "Failed to update Integration status")
			return reconcile.Result{}, err
		}
	}

	// Update status.ServiceName if needed
	serviceName := nameForService(integration)
	if !reflect.DeepEqual(serviceName, integration.Status.ServiceName) {
		integration.Status.ServiceName = serviceName
		err := r.client.Status().Update(context.TODO(), integration)
		if err != nil {
			reqLogger.Error(err, "Failed to update Integration status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: true}, nil
}
