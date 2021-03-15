/*
 * Copyright (c) 2021 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
	"k8s.io/api/extensions/v1beta1"
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
	"time"
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
	err = c.Watch(&source.Kind{Type: &wso2v1alpha2.Integration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Uncomment below if you configure reconcile logic to cater with secondary resources watch
	// Watch for changes to secondary resource Pods and requeue the owner Integration
	// Watch for deployment
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &wso2v1alpha2.Integration{},
	//})
	//
	//// Watch for service
	//err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &wso2v1alpha2.Integration{},
	//})
	//
	//if err != nil {
	//	return err
	//}

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

	err = r.createOrUpdateHPA(eiConfig)
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

	//set to reconcile again after configured interval
	var reconcileIntervalAsStr = eiConfig.integrationConfigMap.Data[reconcileIntervalKey]
	var reconcileInterval, convErr = strconv.Atoi(reconcileIntervalAsStr)
	if convErr != nil {
		reconcileInterval = 10
	}
	return reconcile.Result{ RequeueAfter: time.Duration(reconcileInterval) * time.Second, Requeue: true}, nil
}

// createOrUpdateDeployment updates the existing deployment, if not create a new one
func (r *ReconcileIntegration) createOrUpdateDeployment(config EIConfigNew) (*appsv1.Deployment, error) {
	deployment := r.deploymentForIntegration(config)
	err := k8s.Apply(&r.client, deployment)		//this call modifies the deployment
	return deployment, err
}

// createOrUpdateHPA Checks if auto scaling is enabled and
//create or update horizontal autoscaler for the deployment
func (r *ReconcileIntegration) createOrUpdateHPA(config EIConfigNew) error {
	var autoScaleEnabled bool
	autoScaleEnabled, _ = strconv.ParseBool(config.integration.Spec.AutoScale.Enabled)
	if autoScaleEnabled {
		hpa := createIntegrationHPA(config)
		err := k8s.Apply(&r.client, hpa)
		return err
	}
	return nil
}

// createOrUpdateService Creates or updates k8s service for the deployment
func (r *ReconcileIntegration) createOrUpdateService(config EIConfigNew) error {
	service := &corev1.Service{}
	var integration = config.integration
	namespace := types.NamespacedName{Name: nameForService(&integration), Namespace: integration.Namespace}
	err := k8s.Get(&r.client, namespace, service)
	if err != nil {
		if errors.IsNotFound(err) {
			//service not found, create it
			serviceFromConfig := r.serviceForIntegration(config)
			err := k8s.Create(&r.client,serviceFromConfig)
			if err != nil {
				log.Error(err, "Error creating k8s service", "object", serviceFromConfig)
				return err
			}
			log.Info("Creating k8s service is success", "kind",
				serviceFromConfig.Kind, "object", serviceFromConfig)
		} else {
			log.Error(err, "Failed to get Service for integration")
			return err
		}
	} else {
		//service exists, update it
		//workaround for https://github.com/kubernetes/kubernetes/issues/36072
		serviceFromConfig := r.serviceForIntegration(config)
		serviceFromConfig.ObjectMeta.ResourceVersion = service.ObjectMeta.ResourceVersion
		serviceFromConfig.Spec.ClusterIP = service.Spec.ClusterIP
		err := k8s.Apply(&r.client,serviceFromConfig)
		if err != nil {
			log.Error(err, "Failed to update Service for integration " ,
				"serviceName", serviceFromConfig.Name)
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
		namespace := types.NamespacedName{Name: nameForIngress(), Namespace: integration.Namespace}
		err := k8s.Get(&r.client, namespace, ingress)
		if err != nil {
			if errors.IsNotFound(err) {		// No ingress found, define a new Ingress
				eiIngress := r.ingressForIntegration(config)
				err = k8s.Apply(&r.client, eiIngress)
				if err != nil {
					return err
				}
			} else {
				log.Error(err, "Failed to get Ingress")
				return err
			}

		} else {		//  ingress already exists, check and update the rules
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
				}
			}
		}
	}
	return nil
}

//updateStatus updates the status of the integration
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

	// Update status.ServiceName if needed
	serviceName := nameForService(&config.integration)
	if !reflect.DeepEqual(serviceName, config.integration.Status.ServiceName) {
		config.integration.Status.ServiceName = serviceName
		err := r.client.Status().Update(context.TODO(), &config.integration)
		return err
	}
	return nil
}


