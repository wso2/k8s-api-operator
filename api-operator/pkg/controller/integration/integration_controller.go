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
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
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
	c, err := controller.New(integrationControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Integration
	err = c.Watch(&source.Kind{Type: &wso2v1alpha1.Integration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Integration
	// Watch for deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.Integration{},
	})

	// Watch for service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.Integration{},
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
// Note: The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIntegration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Integration")

	// Fetch the Integration integration
	integration := &wso2v1alpha1.Integration{}
	err := r.client.Get(context.TODO(), request.NamespacedName, integration)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Error fetching the integration object.")
		return reconcile.Result{}, err
	}

	//populate configurations
	eiConfig, configErr := r.PopulateConfigurations(integration)
	if configErr != nil {return reconcile.Result{}, err}

	//create or update the deployment
	deploymentObj, err := r.createOrUpdateDeployment(eiConfig)
	if err != nil {
		reqLogger.Error(err, "Failed to create or update the deployment.",
			"Deployment.Namespace", eiConfig.integration.Namespace,
			"Deployment.Name", eiConfig.integration.Name)
		return reconcile.Result{}, err
	}

	err = r.createOrUpdateHPA(*deploymentObj, eiConfig)
	if err != nil {
		reqLogger.Info("Failed to create/update HPA for the deployment.",
			"Integration.Namespace", integration.Namespace, "Integration.Name", integration.Name)
		return reconcile.Result{}, err
	}

	err = r.createOrUpdateService(eiConfig)
	if err != nil {
		reqLogger.Info("Failed to create/update service for the deployment",
			"Integration.Namespace", integration.Namespace, "Integration.Name", integration.Name)
		return reconcile.Result{}, err
	}

	//create or update ingress
	err = r.createOrUpdateIngress(&eiConfig)
	if err != nil {
		reqLogger.Info("Failed to create/update ingress for the deployment",
			"Integration.Namespace", integration.Namespace, "Integration.Name", integration.Name)
		return reconcile.Result{}, err
	}

	//update status
	err = r.updateStatus(deploymentObj, eiConfig)
	if err != nil {
		reqLogger.Error(err, "Failed to update Integration status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileIntegration) createOrUpdateDeployment(config EIConfigNew) (*appsv1.Deployment, error) {
	// Check if the deployment already exists, if not create a new one
	//deploymentObj := &appsv1.Deployment{}
	//var integration = config.integration
	//namespace := types.NamespacedName{Name: nameForDeployment(&integration), Namespace: integration.Namespace}
	//err := k8s.Get(&r.client, namespace, deploymentObj)
	// Define a new deployment
	deployment := r.deploymentForIntegration(config)
	err := k8s.Apply(&r.client, deployment)
	if err != nil {
		return deployment, err
	}
	//if err != nil && errors.IsNotFound(err) {
	//	log.Info("Creating a new Deployment", "Deployment.Namespace",
	//		deployment.Namespace, "Deployment.Name", deployment.Name)
	//} else if err != nil {
	//	return deploymentObj, err
	//}
	return deployment, nil
}

// createOrUpdateHPA Checks if auto scaling is enabled and
//create or update horizontal autoscaler for the deployment
func (r *ReconcileIntegration) createOrUpdateHPA(deploymentObj appsv1.Deployment, config EIConfigNew) error {
	var autoScaleEnabled bool
	autoScaleEnabled, _ = strconv.ParseBool(config.integration.Spec.AutoScale.Enabled)
	if autoScaleEnabled {
		hpa := createIntegrationHPA(deploymentObj, config)
		// create or update HPA
		err := k8s.Apply(&r.client, hpa)
		return err
	}
	return nil
}

// createOrUpdateService Creates or updates k8s service
func (r *ReconcileIntegration) createOrUpdateService(config EIConfigNew) error {
	// Check if the service already exists, if not create a new one
	serviceObj := &corev1.Service{}
	var integration = config.integration
	namespace := types.NamespacedName{Name: nameForDeployment(&integration), Namespace: integration.Namespace}
	err := k8s.Get(&r.client, namespace, serviceObj)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		service := r.serviceForIntegration(config)
		//reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = k8s.Apply(&r.client, service)
		if err != nil {
			//reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return err
		}
		// Service created successfully - return and requeue
		//return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		//reqLogger.Error(err, "Failed to get Service")
		return err
	} else {			//TODO: check if we need to handle via k8s.Apply
		// Update status.ServiceName if needed
		serviceName := nameForService(&config.integration)
		if !reflect.DeepEqual(serviceName, integration.Status.ServiceName) {
			integration.Status.ServiceName = serviceName
			err := r.client.Status().Update(context.TODO(), &config.integration)
			return err
		}
	}

	return nil
}

// createOrUpdateIngress check if the ingress already exists, if not create a new one, if yes update it
func (r *ReconcileIntegration) createOrUpdateIngress(config *EIConfigNew) error {
	autoCreateIngressInfo := config.integrationConfigMap.Data[autoIngressCreationKey]
	autoCreateIngress, err := strconv.ParseBool(autoCreateIngressInfo)
	if err != nil {
		log.Error(err, "Cannot parse autoIngressCreationKey to a boolean value. Setting false")
		autoCreateIngress = false
	}

	if autoCreateIngress {
		ingress := &v1beta1.Ingress{}
		var integration = config.integration
		namespace := types.NamespacedName{Name: nameForDeployment(&integration), Namespace: integration.Namespace}
		err := k8s.Get(&r.client, namespace, ingress)
		if err != nil {
			if errors.IsNotFound(err) {
				// Define a new Ingress
				eiIngress := r.ingressForIntegration(config)
				//reqLogger.Info("Creating a new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
				err = k8s.Apply(&r.client, eiIngress)
				if err != nil {
					//reqLogger.Error(err, "Failed to create new Ingress", "Ingress.Namespace", integration.Namespace, "Ingress.Name", nameForIngress())
					return err
				}
				// Ingress created successfully - return and requeue
				//reqLogger.Info("Ingress created successfully")
			} else {
				log.Error(err, "Failed to get Ingress")
				return err
				//return reconcile.Result{}, err
			}

		} else {
			_, ruleExists := CheckIngressRulesExist(config, ingress)
			if !ruleExists {
				eiIngress := r.updateIngressForIntegration(config, ingress)
				log.Info("Updating a new Ingress", "Ingress.Namespace", integration.Namespace,
					"Ingress.Name", nameForIngress())
				err = r.client.Update(context.TODO(), eiIngress)
				if err != nil {
					log.Error(err, "Failed to updated new Ingress", "Ingress.Namespace",
						integration.Namespace, "Ingress.Name", nameForIngress())
					return err
					//return reconcile.Result{}, err
				}
				// Ingress updated successfully - return and requeue
				//reqLogger.Info("Ingress updated successfully")
			}
		}
	}
	return nil
}

func (r *ReconcileIntegration) updateStatus(deploymentObj *appsv1.Deployment, config EIConfigNew) error {
	// Update status.Status if needed
	availableReplicas := deploymentObj.Status.AvailableReplicas
	currentStatus := "NotRunning"
	if availableReplicas > 0 {
		currentStatus = "Running"
	}
	if !reflect.DeepEqual(currentStatus, config.integration.Status.Readiness) {
		config.integration.Status.Readiness = currentStatus
		err := r.client.Status().Update(context.TODO(), &config.integration)
		return err
	}
	return nil
}





//TODO: do we need this
func verifyAndUpdateDeployment(config EIConfigNew) error {
	// Ensure the deployment replicas is the same as the spec
	//replicas := integration.Spec.DeploySpec.MinReplicas
	//if *deploymentObj.Spec.Replicas != replicas {
	//	deploymentObj.Spec.Replicas = &replicas
	//	err = r.client.Update(context.TODO(), deploymentObj)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", deploymentObj.Namespace, "Deployment.Name", deploymentObj.Name)
	//		return reconcile.Result{}, err
	//	}
	//	// Spec updated - return and requeue
	//	return reconcile.Result{Requeue: true}, nil
	//}
	return nil
}


