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
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"

	v1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/serving/v1alpha1"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
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
		reqLogger.Info("Operator was not found in the " + instance.Namespace + " namespace. No owner will be set for the artifacts")
	}
	//get configurations file for the controller
	controlConf, err := getConfigmap(r, "controller-config", config.SystemNamespace)
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
	getResourceReqCPU = controlConfigData[resourceRequestCPUTarget]
	getResourceReqMemory := controlConfigData[resourceRequestMemoryTarget]
	getResourceLimitCPU := controlConfigData[resourceLimitCPUTarget]
	getResourceLimitMemory := controlConfigData[resourceLimitMemoryTarget]

	var reqCpu string
	if instance.Spec.Deploy.ReqCpu != "" {
		reqCpu = instance.Spec.Deploy.ReqCpu
	} else {
		reqCpu = getResourceReqCPU
	}
	var reqMemory string
	if instance.Spec.Deploy.ReqMemory != "" {
		reqMemory = instance.Spec.Deploy.ReqMemory
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
		limitMemory = instance.Spec.Deploy.MemoryLimit
	} else {
		limitMemory = getResourceLimitMemory
	}

	var mode string
	mode = instance.Spec.Mode.String()

	if mode == "" {
		mode = privateJet
	}

	minReplicas := int32(instance.Spec.Deploy.MinReplicas)
	if minReplicas <= 0 {
		minReplicas = 1
	}

	if instance.Spec.Deploy.DockerImage != "" && strings.EqualFold(mode, serverless) {
		if err := r.reconcileKnativeDeployment(instance); err != nil {
			return reconcile.Result{}, err
		}
	} else if instance.Spec.Deploy.DockerImage != "" && strings.EqualFold(mode, privateJet) {

		reqLogger.Info("Reconcile K8s Endpoint")
		if err := r.reconcileDeployment(instance, reqCpu, reqMemory, limitCpu, limitMemory, minReplicas); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.reconcileService(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	dep := r.newDeploymentForCR(instance, reqCpu, reqMemory, limitCpu, limitMemory, minReplicas)
	if strings.EqualFold(mode, privateJet) {
		errHpa := createHPA(&r.client, instance, dep, minReplicas, owner)
		if errHpa != nil {
			log.Error(errHpa, "Error creating HPA")
		}
	}

	return reconcile.Result{}, nil
}

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileTargetEndpoint) newDeploymentForCR(m *wso2v1alpha1.TargetEndpoint, resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string, minReplicas int32) *appsv1.Deployment {

	replicas := int32(minReplicas)

	req := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceReqCPU),
		corev1.ResourceMemory: resource.MustParse(resourceReqMemory),
	}
	lim := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceLimitCPU),
		corev1.ResourceMemory: resource.MustParse(resourceLimitMemory),
	}

	// set container ports
	containerPorts := make([]corev1.ContainerPort, 0, len(m.Spec.Ports))
	for _, port := range m.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
		})
	}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       deploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: m.ObjectMeta.Namespace,
			Labels:    m.Labels,
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
						Ports: containerPorts,
						Resources: corev1.ResourceRequirements{
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
	// set container ports
	containerPorts := make([]corev1.ContainerPort, 0, len(m.Spec.Ports))
	for _, port := range m.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
		})
	}

	ser := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: knativeApiVersion,
			Kind:       serviceKind,
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
						Annotations: map[string]string{
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
									Ports: containerPorts,
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

func (r *ReconcileTargetEndpoint) reconcileDeployment(m *wso2v1alpha1.TargetEndpoint, resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string, minReplicas int32) error {

	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeploymentForCR(m, resourceReqCPU, resourceReqMemory, resourceLimitCPU, resourceLimitMemory, minReplicas)
		log.WithValues("Creating a new Deployment [namespace] ", dep.Namespace, "[deployment-name]", dep.Name)

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
		log.WithValues("Creating new Knative Service [namespace] ", ser.Namespace, " [knative-service-name] ", ser.Name)

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
	// set service ports
	servicePorts := make([]corev1.ServicePort, 0, len(m.Spec.Ports))
	for _, port := range m.Spec.Ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
		})
	}

	protocol := m.Spec.ApplicationProtocol
	port := m.Spec.Ports[0].Port
	targetPort := int(m.Spec.Ports[0].TargetPort)

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
	servicePorts[0].Port = port
	servicePorts[0].TargetPort = intstr.FromInt(targetPort)

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       serviceKind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: m.ObjectMeta.Namespace,
			Labels:    m.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: m.ObjectMeta.Labels,
			Ports:    servicePorts,
		},
	}
	controllerutil.SetControllerReference(m, &service, r.scheme)
	return &service
}

// createHPA checks whether the HPA version is v2beta1 or v2beta2
func createHPA(client *client.Client, targetEp *wso2v1alpha1.TargetEndpoint, dep *appsv1.Deployment, minReplicas int32,
	owner []metav1.OwnerReference) error {
	// get global hpa configs, return error if not found (required config map)
	hpaConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: hpaConfigMapName}, hpaConfMap)
	if err != nil {
		return err
	}
	if hpaConfMap.Data[hpaVersionConst] == "v2beta1" {
		hpaV2beta1, errHpaV2beta1 := createHPAv2beta1(client, targetEp, dep, minReplicas, owner)
		if errHpaV2beta1 != nil {
			return errHpaV2beta1
		}
		//create or apply HPA
		return k8s.Apply(client, hpaV2beta1)
	}
	if hpaConfMap.Data[hpaVersionConst] == "v2beta2" {
		hpaV2beta2, errHpaV2beta2 := createHPAv2beta2(client, targetEp, dep, minReplicas, owner)
		if errHpaV2beta2 != nil {
			return errHpaV2beta2
		}
		//create or apply HPA
		return k8s.Apply(client, hpaV2beta2)
	}
	return err
}

// createHPA creates (or update) HPA for the Target Endpoint with HPA version v2beta1
func createHPAv2beta1(client *client.Client, targetEp *wso2v1alpha1.TargetEndpoint, dep *appsv1.Deployment, minReplicas int32,
	owner []metav1.OwnerReference) (*v2beta1.HorizontalPodAutoscaler, error) {
	// target resource
	targetResource := v2beta1.CrossVersionObjectReference{
		Kind:       dep.Kind,
		Name:       dep.Name,
		APIVersion: dep.APIVersion,
	}

	// get global hpa configs, return error if not found (required config map)
	hpaConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: hpaConfigMapName}, hpaConfMap)
	if err != nil {
		log.Error(err, "HPA configs not defined")
		return nil, err
	}

	// setting max replicas
	maxReplicas := targetEp.Spec.Deploy.MaxReplicas
	if maxReplicas <= 0 {
		// setting default max replicas from the configmap "hpa-config"
		maxReplicasInt64, errInt := strconv.ParseInt(hpaConfMap.Data[maxReplicasConfigKey], 10, 32)
		if errInt != nil {
			log.Error(err, "Error parsing HPA MaxReplicas",
				"value", hpaConfMap.Data[maxReplicasConfigKey])
			return nil, err
		}
		maxReplicas = int32(maxReplicasInt64)
	}

	// parse hpa config yaml
	var metricsHpa []v2beta1.MetricSpec
	yamlErr := yaml.Unmarshal([]byte(hpaConfMap.Data[metricsConfigKeyV2beta1]), &metricsHpa)
	if yamlErr != nil {
		log.Error(err, "Error marshalling HPA config yaml", "configmap", hpaConfMap)
		return nil, yamlErr
	}

	// HPA instance for Target Endpoint
	hpa := &v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dep.Name,
			Namespace:       dep.Namespace,
			OwnerReferences: owner,
		},
		Spec: v2beta1.HorizontalPodAutoscalerSpec{
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        metricsHpa,
		},
	}

	return hpa, nil
}

// createHPA creates (or update) HPA for the Target Endpoint with HPA version v2beta1
func createHPAv2beta2(client *client.Client, targetEp *wso2v1alpha1.TargetEndpoint, dep *appsv1.Deployment, minReplicas int32,
	owner []metav1.OwnerReference) (*v2beta2.HorizontalPodAutoscaler, error) {
	// target resource
	targetResource := v2beta2.CrossVersionObjectReference{
		Kind:       dep.Kind,
		Name:       dep.Name,
		APIVersion: dep.APIVersion,
	}

	// get global hpa configs, return error if not found (required config map)
	hpaConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: hpaConfigMapName}, hpaConfMap)
	if err != nil {
		return nil, err
	}

	// setting max replicas
	maxReplicas := targetEp.Spec.Deploy.MaxReplicas
	if maxReplicas <= 0 {
		// setting default max replicas from the configmap "hpa-config"
		maxReplicasInt64, errInt := strconv.ParseInt(hpaConfMap.Data[maxReplicasConfigKey], 10, 32)
		if errInt != nil {
			log.Error(err, "Error parsing HPA MaxReplicas",
				"value", hpaConfMap.Data[maxReplicasConfigKey])
			return nil, err
		}
		maxReplicas = int32(maxReplicasInt64)
	}

	// parse hpa config yaml
	var metricsHpa []v2beta2.MetricSpec
	yamlErr := yaml.Unmarshal([]byte(hpaConfMap.Data[metricsConfigKey]), &metricsHpa)
	if yamlErr != nil {
		log.Error(err, "Error marshalling HPA config yaml", "configmap", hpaConfMap)
		return nil, yamlErr
	}

	// HPA instance for Target Endpoint
	hpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dep.Name,
			Namespace:       dep.Namespace,
			OwnerReferences: owner,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        metricsHpa,
		},
	}

	return hpa, nil
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
