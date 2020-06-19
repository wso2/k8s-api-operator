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
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/analytics"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/endpoints"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/interceptors"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ratelimit"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/security"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/swagger"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
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
	"strconv"
	"strings"
	"time"
)

var log = logf.Log.WithName("api.controller")

// Add creates a new API Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAPI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource API
	err = c.Watch(&source.Kind{Type: &wso2v1alpha1.API{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner API
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.API{},
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
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a API object and makes changes based on the state read
// and what is in the API.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("request_namespace", request.Namespace, "request_name", request.Name)
	reqLogger.Info("Reconciling API")

	// initialize volumes
	kaniko.InitDocFileProp()
	kaniko.InitJobVolumes()
	mgw.InitContainers()

	var apiVersion string // API version - for the tag of final MGW docker image

	apiBasePathMap := make(map[string]string) // API base paths with versions

	// Fetch the API instance
	instance := &wso2v1alpha1.API{}
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

	ownerRef := k8s.NewOwnerRef(instance.TypeMeta, instance.ObjectMeta)
	if errOwnerSet := setApiDependent(&r.client, instance, ownerRef); errOwnerSet != nil {
		log.Error(errOwnerSet, "Error setting owner ref for API dependent configs")
		return reconcile.Result{}, errOwnerSet
	}

	operatorOwner, ownerErr := getOperatorOwner(&r.client)
	if ownerErr != nil {
		reqLogger.Info("Operator was not found in the " + wso2NameSpaceConst + " namespace. No owner will be set for the artifacts")
	}
	userNamespace := instance.Namespace

	//get configurations file for the controller
	controlConf := k8s.NewConfMap()
	errConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: controllerConfName}, controlConf)
	//get docker registry configs
	dockerRegistryConf := k8s.NewConfMap()
	errRegConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: dockerRegConfigs}, dockerRegistryConf)

	confErrs := []error{errConf, errRegConf}
	for _, err := range confErrs {
		if err != nil {
			if errors.IsNotFound(err) {
				// Required configmap is not found, could have been deleted after reconcile request.
				// Return and requeue
				log.Error(err, "Required configmap is not found")
				return reconcile.Result{}, err
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	}

	controlConfigData := controlConf.Data
	kaniko.DocFileProp.ToolkitImage = controlConfigData[mgwToolkitImgConst]
	kaniko.DocFileProp.RuntimeImage = controlConfigData[mgwRuntimeImgConst]

	mgwDockerImage := registry.Image{}
	registryTypeStr := dockerRegistryConf.Data[registryTypeConst]
	if !registry.IsRegistryType(registryTypeStr) {
		log.Error(err, "Invalid registry type", "registry-type", registryTypeStr)
		// Registry type is invalid, user should update this with valid type.
		// Return and requeue
		return reconcile.Result{}, err
	}
	mgwDockerImage.RegistryType = registry.Type(dockerRegistryConf.Data[registryTypeConst])
	mgwDockerImage.RepositoryName = dockerRegistryConf.Data[repositoryNameConst]
	operatorMode := controlConfigData[operatorModeConst]

	// log controller configurations
	reqLogger.Info(
		"Controller configurations",
		"mgw_toolkit_image", kaniko.DocFileProp.ToolkitImage, "mgw_runtime_image", kaniko.DocFileProp.RuntimeImage,
		"gateway_observability", controlConfigData[observabilityEnabledConfigKey],
		"user_nameSpace", userNamespace, "operator_mode", operatorMode,
	)
	// log registry configurations
	reqLogger.Info(
		"Registry configurations", "registry_type", mgwDockerImage.RegistryType,
		"repository_name", mgwDockerImage.RepositoryName,
	)

	// validate HPA configs and setting configs
	// this is to verify HPA configs prior running kaniko job and creating MGW image
	// otherwise user may have to wait long time to know the error in configs
	mgw.Configs.ObservabilityEnabled = strings.EqualFold(controlConfigData[observabilityEnabledConfigKey], "true")
	if err := mgw.ValidateHpaConfigs(&r.client); err != nil {
		// error has already logged inside the method
		// Return and requeue request since config mismatch
		return reconcile.Result{}, err
	}

	// if there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies
	if err := ratelimit.Handle(&r.client, userNamespace, operatorOwner); err != nil {
		log.Error(err, "Error in creating default policy configmap")
	}

	swaggerCmNames := instance.Spec.Definition.SwaggerConfigmapNames
	for i, swaggerCmName := range swaggerCmNames {
		// Check if the configmap mentioned in the crd object exist
		swaggerConfMap := k8s.NewConfMap()
		err := k8s.Get(&r.client, types.NamespacedName{Namespace: userNamespace, Name: swaggerCmName}, swaggerConfMap)
		if err != nil {
			if errors.IsNotFound(err) {
				// Swagger configmap is not found, could have been deleted after reconcile request.
				// Return and requeue
				return reconcile.Result{}, err
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}

		// fetch swagger data from configmap and loads as open api swagger v3
		swaggerFileName, errSwagger := maps.OneKey(swaggerConfMap.Data)
		if errSwagger != nil {
			log.Error(errSwagger, "Error in the swagger configmap data", "data", swaggerConfMap.Data)
			return reconcile.Result{}, errSwagger
		}
		swaggerData := swaggerConfMap.Data[swaggerFileName]
		swaggerFileName = str.GetRandFileName(swaggerFileName) // randomize file name to make it unique
		swaggerDoc, err := swagger.GetSwaggerV3(&swaggerData)
		if err != nil {
			return reconcile.Result{}, nil
		}

		// Set endpoint deployment mode: sidecar/private-jet
		epDeployMode, errMode := swagger.EpDeployMode(instance, swaggerDoc)
		if errMode != nil {
			log.Error(errMode, "Error setting the endpoint deployment mode")
			return reconcile.Result{}, errMode
		}

		apiVersion = swaggerDoc.Info.Version
		endpointNames := swagger.HandleMgwEndpoints(&r.client, swaggerDoc, epDeployMode, userNamespace)
		apiBasePath := swagger.ApiBasePath(swaggerDoc)
		apiBasePathMap[apiBasePath] = apiVersion

		// Creating sidecar endpoint deployment
		if epDeployMode == sidecar {
			err := endpoints.AddSidecarContainers(&r.client, userNamespace, &endpointNames)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		reqLogger.Info("getting security instance")
		//check security scheme already exist
		_, secSchemeDefined := swaggerDoc.Extensions[swagger.SecuritySchemeExtension]

		securityMap, isDefinedSecurity, resourceLevelSec, securityErr := swagger.GetSecurityMap(swaggerDoc)
		if securityErr != nil {
			return reconcile.Result{}, securityErr
		}

		securityDefinition, jwtConfArray, errSec := security.Handle(&r.client, securityMap, userNamespace, secSchemeDefined)
		if errSec != nil {
			return reconcile.Result{}, errSec
		}
		mgw.Configs.JwtConfigs = jwtConfArray
		//adding security scheme to swagger
		if len(securityDefinition) > 0 {
			swaggerDoc.Components.Extensions[swagger.SecuritySchemeExtension] = securityDefinition
		}

		// mount formatted swagger to kaniko job
		formattedSwagger := swagger.PrettyString(swaggerDoc)
		formattedSwaggerCmName := swaggerCmName + "-mgw"
		// create configmap with modified swagger
		swaggerDataMgw := map[string]string{swaggerFileName: formattedSwagger}
		swaggerConfMapMgw := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNamespace, Name: formattedSwaggerCmName}, &swaggerDataMgw, nil, ownerRef)
		log.Info("Creating swagger configmap for mgw", "name", formattedSwaggerCmName, "namespace", userNamespace)
		// add to kaniko volumes
		if instance.Spec.UpdateTimeStamp == "" {
			if err := k8s.CreateIfNotExists(&r.client, swaggerConfMapMgw); err != nil {
				log.Error(err, "Error creating formatted swagger configmap", "configmap", swaggerConfMapMgw)
				return reconcile.Result{}, err
			}
		} else {
			log.Info("Updating swagger-mgw since timestamp value is given")
			if err := k8s.Update(&r.client, swaggerConfMapMgw); err != nil {
				log.Error(err, "Error updating formatted swagger configmap", "configmap", swaggerConfMapMgw)
				return reconcile.Result{}, err
			}
		}
		kaniko.AddVolume(k8s.ConfigMapVolumeMount(swaggerConfMapMgw.Name, fmt.Sprintf(kaniko.SwaggerLocation, i+1)))

		// Default security
		if !isDefinedSecurity && resourceLevelSec == 0 {
			log.Info("Use default security")

			err := security.Default(&r.client, userNamespace, ownerRef)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	// micro-gateway image to be build
	mgwDockerImage.Name = strings.ToLower(strings.ReplaceAll(instance.Name, " ", ""))
	mgwDockerImage.Tag = apiVersion
	// if multi swagger mode override image tag
	if len(swaggerCmNames) > 1 {
		if instance.Spec.Version != "" {
			mgwDockerImage.Tag = instance.Spec.Version
		} else {
			// if not defined in the API Crd set default
			mgwDockerImage.Tag = apiCrdDefaultVersion
		}
	}
	if instance.Spec.UpdateTimeStamp != "" {
		mgwDockerImage.Tag = mgwDockerImage.Tag + "-" + instance.Spec.UpdateTimeStamp
	}

	errReg := registry.SetRegistry(&r.client, userNamespace, mgwDockerImage)
	if errReg != nil {
		log.Error(errReg, "Error setting docker registry")
		return reconcile.Result{}, errReg
	}

	// check if the image already exists
	imageExist, errImage := registry.IsImageExist(&r.client)
	if errImage != nil {
		log.Error(errImage, "Error in image finding")
	}
	log.Info("Is MGW runtime image exist in the docker registry?",
		"exists", strconv.FormatBool(imageExist), "mgw_docker_image", mgwDockerImage)

	// handling analytics
	log.Info("Handling analytics")
	if err := analytics.Handle(&r.client, userNamespace); err != nil {
		return reconcile.Result{}, err
	}

	// handling interceptors
	log.Info("Handling interceptors")
	if err := interceptors.Handle(&r.client, instance); err != nil {
		log.Error(err, "Error handling interceptors")
		return reconcile.Result{}, err
	}

	// handling Kaniko docker file
	log.Info("Rendering the dockerfile for Kaniko job and adding volumes to the Kaniko job")
	if err := kaniko.HandleDockerFile(&r.client, userNamespace, instance.Name, ownerRef); err != nil {
		log.Error(err, "Error rendering the docker file for Kaniko job and adding volumes to the Kaniko job")
		return reconcile.Result{}, err
	}

	// setting the MGW configs from APIM configmap
	log.Info("Setting the MGW configs from APIM configmap")
	if err := mgw.SetApimConfigs(&r.client); err != nil {
		log.Error(err, "Error Setting the MGW configs from APIM configmap")
		return reconcile.Result{}, err
	}

	// rendering MGW config file
	log.Info("Rendering and adding the MGW configuration file to cluster")
	if err := mgw.ApplyConfFile(&r.client, userNamespace, instance.Name, ownerRef); err != nil {
		log.Error(err, "Error rendering and adding the MGW configuration file to cluster")
		return reconcile.Result{}, err
	}

	kanikoArgs := k8s.NewConfMap()
	err = k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: kanikoArgsConfigs}, kanikoArgs)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No kaniko-arguments config map is available in wso2-system namespace")
	}

	// create Kaniko job
	// if updating api or overriding api or image not found
	var kanikoJob *batchv1.Job
	if instance.Spec.UpdateTimeStamp != "" || instance.Spec.Override || !imageExist {
		log.Info("Deploying the Kaniko job in cluster")
		kanikoJob = kaniko.Job(instance, controlConfigData, kanikoArgs.Data[kanikoArguments], ownerRef)
		if err := controllerutil.SetControllerReference(instance, kanikoJob, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		// create Kaniko job and set kaniko object
		if errJob := k8s.CreateIfNotExists(&r.client, kanikoJob); errJob != nil {
			return reconcile.Result{}, errJob
		}
	}
	// if kaniko job started (i.e. not nil)
	if kanikoJob != nil {
		// check for kaniko completion
		for t := 30; kanikoJob.Status.Succeeded == 0 && t > 0; t -= 5 {
			reqLogger.Info("Kaniko job is still not completed", "retry_interval_seconds", t, "job_status", kanikoJob.Status)
			time.Sleep(time.Duration(t * 1e9)) // t seconds (i.e. t * 1e9 nano seconds)
			_ = k8s.Get(&r.client, types.NamespacedName{Namespace: kanikoJob.Namespace, Name: kanikoJob.Name}, kanikoJob)
		}

		if kanikoJob.Status.Succeeded == 0 {
			reqLogger.Info("Kaniko job is still not completed and requeue request after 10 seconds", "job_status", kanikoJob.Status)
			return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		} else {
			reqLogger.Info("Kaniko job is completed successfully", "job_status", kanikoJob.Status)
		}
	}

	// kaniko job completed or not ran (i.e. image already exists)
	// deploying the MGW runtime image
	generateK8sArtifactsForMgw := controlConfigData[generateKubernetesArtifactsForMgw]
	deployMgwRuntime, err := strconv.ParseBool(generateK8sArtifactsForMgw)
	if err != nil {
		log.Error(err, "Error reading value for generate k8s artifacts")
		return reconcile.Result{}, err
	}
	if deployMgwRuntime {
		reqLogger.Info("Deploying MGW runtime image")
		// create MGW deployment in k8s cluster
		mgwDeployment := mgw.Deployment(instance, controlConfigData, ownerRef)
		if errMgw := k8s.Apply(&r.client, mgwDeployment); errMgw != nil {
			reqLogger.Error(errMgw, "Error updating the MGW deployment", "deploy_name", mgwDeployment.Name)
			return reconcile.Result{}, errMgw
		}
		reqLogger.Info("Updated the MGW deployment", "deploy_name", mgwDeployment.Name)

		// create MGW service
		mgwSvc := mgw.Service(instance, operatorMode, *ownerRef)
		// controllerutil.SetControllerReference(instance, mgwSvc, r.scheme) <- check with commenting this, if work delete this.
		if errMgwSvc := k8s.CreateIfNotExists(&r.client, mgwSvc); errMgwSvc != nil {
			reqLogger.Error(errMgwSvc, "Error creating the MGW service", "service_name", mgwSvc.Name)
			return reconcile.Result{}, errMgwSvc
		}

		// create horizontal pod auto-scalar
		hpa := mgw.HPA(instance, mgwDeployment, ownerRef)
		if errHpa := k8s.CreateIfNotExists(&r.client, hpa); errHpa != nil {
			reqLogger.Error(errHpa, "Error creating the horizontal pod auto-scalar", "hpa_name", hpa.Name)
			return reconcile.Result{}, errHpa
		}

		reqLogger.Info("Operator mode", "mode", operatorMode)
		if strings.EqualFold(operatorMode, ingressMode) {
			errIng := mgw.ApplyIngressResource(&r.client, instance, apiBasePathMap, ownerRef)
			if errIng != nil {
				reqLogger.Error(errIng, "Error creating the ingress resource")
				return reconcile.Result{}, errIng
			}
		}
		if strings.EqualFold(operatorMode, routeMode) {
			rutErr := mgw.ApplyRouteResource(&r.client, instance, apiBasePathMap, ownerRef)
			if rutErr != nil {
				return reconcile.Result{}, rutErr
			}
		}

		reqLogger.Info("Successfully deployed the API", "api_name", instance.Name)
		return reconcile.Result{}, nil
	} else {
		reqLogger.Info("Skip updating kubernetes artifacts")
		return reconcile.Result{}, nil
	}
}

// setApiDependent sets API owner reference to dependents
func setApiDependent(client *client.Client, api *wso2v1alpha1.API, ownerRef *[]metav1.OwnerReference) error {
	confMap := &corev1.ConfigMap{}
	// swagger configmaps
	confMapNames := api.Spec.Definition.SwaggerConfigmapNames
	confMapNames = append(confMapNames, api.Spec.Definition.Interceptors.Ballerina...)
	confMapNames = append(confMapNames, api.Spec.Definition.Interceptors.Java...)

	for _, confMapName := range confMapNames {
		// get configmap
		err := k8s.Get(client, types.NamespacedName{
			Namespace: api.Namespace,
			Name:      confMapName,
		}, confMap)
		if err != nil {
			log.Error(err, "Error retrieving api dependent configmap", "configmap", confMapName)
			return err
		}

		// set owner ref
		if err := k8s.UpdateOwner(client, ownerRef, confMap); err != nil {
			log.Error(err, "Error updating api owner reference of dependent configmap", "configmap", confMap)
			return err
		}
	}

	return nil
}

// getOperatorOwner returns the owner reference of the operator
func getOperatorOwner(client *client.Client) (*[]metav1.OwnerReference, error) {
	depFound := &appsv1.Deployment{}
	errDeploy := k8s.Get(client, types.NamespacedName{Name: "api_operator", Namespace: wso2NameSpaceConst}, depFound)
	if errDeploy != nil {
		noOwner := []metav1.OwnerReference{}
		return &noOwner, errDeploy
	}

	return k8s.NewOwnerRef(depFound.TypeMeta, depFound.ObjectMeta), nil
}
