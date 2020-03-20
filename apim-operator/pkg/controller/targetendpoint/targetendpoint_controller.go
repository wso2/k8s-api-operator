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

package targetendpoint

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"

	wso2v1alpha1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/wso2/v1alpha1"
	v1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/serving/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

)

var log = logf.Log.WithName("controller_targetendpoint")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TargetEndpoint Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTargetEndpoint{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("targetendpoint-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TargetEndpoint
	err = c.Watch(&source.Kind{Type: &wso2v1alpha1.TargetEndpoint{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TargetEndpoint
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.TargetEndpoint{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTargetEndpoint{}

// ReconcileTargetEndpoint reconciles a TargetEndpoint object
type ReconcileTargetEndpoint struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TargetEndpoint object and makes changes based on the state read
// and what is in the TargetEndpoint.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTargetEndpoint) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TargetEndpoint")

	// Fetch the Endpoint instance
	instance := &wso2v1alpha1.TargetEndpoint{}
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
	//getting owner reference to create HPA for TargetEndPoint
	owner := getOwnerDetails(instance)
	if owner == nil {
		reqLogger.Info("Operator was not found in the "+instance.Namespace+" namespace. No owner will be set for the artifacts")
	}
	//get configurations file for the controller
	controlConf, err := getConfigmap(r, "controller-config", "wso2-system")
	if err != nil {
		if errors.IsNotFound(err) {
			// Controller configmap is not found, could have been deleted after reconcile request.
			// Return and requeue
			log.Error(err, "Controller configuration file is not found")
			return reconcile.Result{}, err
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	controlConfigData := controlConf.Data
	var getResourceReqCPU string
	getResourceReqCPU = controlConfigData["resourceRequestCPUTarget"]
	getResourceReqMemory := controlConfigData["resourceRequestMemoryTarget"]
	getResourceLimitCPU := controlConfigData["resourceLimitCPUTarget"]
	getResourceLimitMemory := controlConfigData["resourceLimitMemoryTarget"]

	var reqCpu string
	if instance.Spec.Deploy.ReqCpu != "" {
		reqCpu =  instance.Spec.Deploy.ReqCpu
	} else {
		reqCpu = getResourceReqCPU
	}
	var reqMemory string
	if instance.Spec.Deploy.ReqMemory != "" {
		reqMemory =  instance.Spec.Deploy.ReqMemory
	} else {
		reqMemory = getResourceReqMemory
	}
	var limitCpu string
	if instance.Spec.Deploy.LimitCpu != "" {
		limitCpu = instance.Spec.Deploy.LimitCpu
	} else {
		limitCpu = getResourceLimitCPU
	}
	var limitMemory string
	if instance.Spec.Deploy.MemoryLimit != "" {
		limitMemory =  instance.Spec.Deploy.MemoryLimit
	} else {
		limitMemory = getResourceLimitMemory
	}

	if instance.Spec.Deploy.DockerImage != "" && instance.Spec.Mode == "Serverless" {
		if err := r.reconcileKnativeDeployment(instance); err != nil {
			return reconcile.Result{}, err
		}
	} else if instance.Spec.Deploy.DockerImage != "" && instance.Spec.Mode == "privateJet" {
		reqLogger.Info("Reconcile K8s Endpoint")
		if err := r.reconcileDeployment(instance,reqCpu,reqMemory,limitCpu,limitMemory);
		err != nil {
			return reconcile.Result{}, err
		}
		if err := r.reconcileService(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	dep := r.newDeploymentForCR(instance,reqCpu,reqMemory,limitCpu,limitMemory)

	getMaxRep := controlConfigData["hpaMaxReplicas"]
	intValueRep, err := strconv.ParseInt(getMaxRep, 10, 32)
	if err != nil {
		log.Error(err, "error getting max replicas")
	}
	maxReplicas := int32(intValueRep)
	GetAvgUtilCPU := controlConfigData["hpaTargetAverageUtilizationCPU"]
	intValueUtilCPU, err := strconv.ParseInt(GetAvgUtilCPU, 10, 32)
	if err != nil {
		log.Error(err, "error getting hpa target average utilization for CPU")
	}
	targetAvgUtilizationCPU := int32(intValueUtilCPU)
	minReplicas := int32(instance.Spec.Deploy.MinReplicas)
	if instance.Spec.Mode != "Serverless" {
		errGettingHpa := createTargetEndPointHPA(dep,r,owner, minReplicas, maxReplicas, targetAvgUtilizationCPU)
		if errGettingHpa != nil {
			log.Error(errGettingHpa, "Error getting HPA")
		}
	}
	return reconcile.Result{Requeue: true}, nil
}

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileTargetEndpoint) newDeploymentForCR(m *wso2v1alpha1.TargetEndpoint, resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string) *appsv1.Deployment {
	replicas := m.Spec.Deploy.MinReplicas
	req := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceReqCPU),
		corev1.ResourceMemory: resource.MustParse(resourceReqMemory),
	}
	lim := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceLimitCPU),
		corev1.ResourceMemory: resource.MustParse(resourceLimitMemory),
	}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: m.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: m.ObjectMeta.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: m.ObjectMeta.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: m.Spec.Deploy.DockerImage,
						Name:  m.Spec.Deploy.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: m.Spec.Port,
						}},
						Resources:corev1.ResourceRequirements{
							Limits:   lim,
							Requests: req,
						},
					}},
				},
			},
		},
	}
	// Set Examplekind instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep

}

// Create newKnativeDeploymentForCR method to create a deployment.
func (r *ReconcileTargetEndpoint) newKnativeDeploymentForCR(m *wso2v1alpha1.TargetEndpoint) *v1.Service {
	ser := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: m.ObjectMeta.Namespace,
		},
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
						Labels: m.ObjectMeta.Labels,
						Annotations:map[string]string{
							"autoscaling.knative.dev/minScale": strconv.Itoa(int(m.Spec.Deploy.MinReplicas)),
							"autoscaling.knative.dev/maxScale": strconv.Itoa(int(m.Spec.Deploy.MaxReplicas)),
						},
					},
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: m.Spec.Deploy.DockerImage,
									Name:  m.Spec.Deploy.Name,
									Ports: []corev1.ContainerPort{{
										ContainerPort: m.Spec.Port,
									}},
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Examplekind instance as the owner and controller
	controllerutil.SetControllerReference(m, ser, r.scheme)
	return ser
}

func (r *ReconcileTargetEndpoint) reconcileService(m *wso2v1alpha1.TargetEndpoint) error {
	newService := r.newServiceForCR(m)

	err := r.client.Create(context.TODO(), newService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Service resource: %v", err)
	}

	if err == nil {
		return nil
	}

	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: newService.Namespace,
		Name: newService.Name}, currentService)

	if err != nil {
		return fmt.Errorf("failed to retrieve Service resource: %v", err)
	}

	if reflect.DeepEqual(currentService.Spec.Ports, newService.Spec.Ports) {
		return nil
	}

	currentService.Spec.Ports = newService.Spec.Ports
	return r.client.Update(context.TODO(), currentService)
}

func (r *ReconcileTargetEndpoint) reconcileDeployment(m *wso2v1alpha1.TargetEndpoint,resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string) error {
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeploymentForCR(m,resourceReqCPU,resourceReqMemory,resourceLimitCPU,resourceLimitMemory)
		log.WithValues("Creating a new Deployment %s/%s\n", dep.Namespace, dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			log.WithValues("Failed to create new Deployment: %v\n", err)
			return err
		}
		// Deployment created successfully - return and requeue
	} else if err != nil {
		log.WithValues("Failed to get Deployment: %v\n", err)
		return err
	}
	return nil
}

func (r *ReconcileTargetEndpoint) reconcileKnativeDeployment(m *wso2v1alpha1.TargetEndpoint) error {
	found := &v1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		//Define new Knative deployment
		ser := r.newKnativeDeploymentForCR(m)
		log.WithValues("Creating new Knative Service %s%s\n", ser.Namespace, ser.Name)
		err = r.client.Create(context.TODO(), ser)
		if err != nil {
			log.WithValues("Failed to create new Knative Service: %v\n", err)
			return err
		}
		// Knative Service created sucessfully - return and requee
	} else if err != nil {
		log.WithValues("Failed to get Knative Service: %\n", err)
		return err
	}
	return nil
}

// NewService assembles the ClusterIP service for the Nginx
func (r *ReconcileTargetEndpoint) newServiceForCR(m *wso2v1alpha1.TargetEndpoint) *corev1.Service {

	protocol := m.Spec.Protocol
	port := int(m.Spec.Port)
	targetPort := int(m.Spec.TargetPort)

	switch protocol {
	case "https":
		if port == 0 {
			port = 443
		}
		if targetPort == 0 {
			targetPort = 443
		}
	case "http":
		if port == 0 {
			port = 80
		}
		if targetPort == 0 {
			targetPort = 80
		}
	}

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: m.ObjectMeta.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: m.ObjectMeta.Labels,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{Port: m.Spec.Port, TargetPort: intstr.FromInt(targetPort)},
			},
		},
	}
	controllerutil.SetControllerReference(m, &service, r.scheme)
	return &service
}


func createTargetEndPointHPA(dep *appsv1.Deployment ,r *ReconcileTargetEndpoint, owner []metav1.OwnerReference,
	minReplicas int32, maxReplicas int32, targetAverageUtilizationCPU int32) error {

	instance := &wso2v1alpha1.TargetEndpoint{}


	targetResource := v2beta1.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       dep.Name,
		APIVersion: "extensions/v1beta1",
	}
	//CPU utilization
	resourceMetricsForCPU := &v2beta1.ResourceMetricSource{
		Name:                     corev1.ResourceCPU,
		TargetAverageUtilization: &targetAverageUtilizationCPU,
	}
	metricsResCPU := v2beta1.MetricSpec{
		Type:     "Resource",
		Resource: resourceMetricsForCPU,
	}
	metricsSet := []v2beta1.MetricSpec{metricsResCPU}
	hpa := &v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dep.Name + "-hpa",
			Namespace:       dep.Namespace,
			OwnerReferences: owner,
		},
		Spec: v2beta1.HorizontalPodAutoscalerSpec{
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        metricsSet,
		},
	}
	//check hpa already exists
	checkHpa := &v2beta1.HorizontalPodAutoscaler{}
	hpaErr := r.client.Get(context.TODO(), types.NamespacedName{Name: hpa.Name, Namespace: hpa.Namespace}, checkHpa)
	if hpaErr != nil && errors.IsNotFound(hpaErr) {
		//creating new hpa
		log.Info("Creating HPA for targetEndPoint " + instance.Name)
		errHpaCreating := r.client.Create(context.TODO(), hpa)
		if errHpaCreating != nil {
			return errHpaCreating
		}
		return nil
	} else if hpaErr != nil {
		return hpaErr
	} else {
		log.Info("HPA for targetEndPoint " + instance.Name + " is already exist")
	}
	return nil
}

//get configmap
func getConfigmap(r *ReconcileTargetEndpoint, mapName string, ns string) (*corev1.ConfigMap, error) {
	apiConfigMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: mapName, Namespace: ns}, apiConfigMap)

	if mapName == "apim-config" {
		if err != nil && errors.IsNotFound(err) {
			logrus.Warnf("missing APIM configurations ", err)
			return nil, err

		} else if err != nil {
			log.Error(err, "error ")
			return apiConfigMap, err
		}
	} else {
		if err != nil && errors.IsNotFound(err) {
			log.Error(err, "Specified configmap is not found: %s", mapName)
			return apiConfigMap, err
		} else if err != nil {
			log.Error(err, "error ")
			return apiConfigMap, err
		}
	}
	return apiConfigMap, nil
}

//gets the details of the targetEndPoint crd object for owner reference
func getOwnerDetails(cr *wso2v1alpha1.TargetEndpoint) []metav1.OwnerReference {
	setOwner := true
	return []metav1.OwnerReference{
		{
			APIVersion:         cr.APIVersion,
			Kind:               cr.Kind,
			Name:               cr.Name,
			UID:                cr.UID,
			Controller:         &setOwner,
			BlockOwnerDeletion: &setOwner,
		},
	}
}
