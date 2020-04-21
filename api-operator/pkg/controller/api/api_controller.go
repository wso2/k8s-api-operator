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
	"github.com/golang/glog"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/analytics"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/interceptors"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/security"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	swagger "github.com/wso2/k8s-api-operator/api-operator/pkg/swagger"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/volume"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/intstr"

	routv1 "github.com/openshift/api/route/v1"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
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

	"encoding/json"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/ratelimiting"
)

var log = logf.Log.WithName("api.controller")

//These structs used to build the security schema in json
type path struct {
	Security []map[string][]string `json:"security"`
}
type securitySchemeStruct struct {
	SecurityType string             `json:"type"`
	Scheme       string             `json:"scheme,omitempty"`
	Flows        *authorizationCode `json:"flows,omitempty"`
}

type SecurityTypeJWT struct {
	CertificateAlias     string
	Issuer               string
	Audience             string
	ValidateSubscription bool
}
type authorizationCode struct {
	AuthorizationCode scopeSet `json:"authorizationCode,omitempty"`
}
type scopeSet struct {
	AuthorizationUrl string            `json:"authorizationUrl"`
	TokenUrl         string            `json:"tokenUrl"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

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

var kanikoArgs = &corev1.ConfigMap{}

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
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling API")

	// initialize volumes
	kaniko.InitDocFileProp()
	volume.InitJobVolumes()

	var apiVersion string // API version - for the tag of final MGW docker image

	apiBasePathMap := make(map[string]string) // API base paths with versions

	// Fetch the API instance
	instance := &wso2v1alpha1.API{}

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

	owner := k8s.NewOwnerRef(instance.TypeMeta, instance.ObjectMeta)
	operatorOwner, ownerErr := getOperatorOwner(r)
	if ownerErr != nil {
		reqLogger.Info("Operator was not found in the " + wso2NameSpaceConst + " namespace. No owner will be set for the artifacts")
	}
	userNamespace := instance.Namespace

	//get configurations file for the controller
	controlConf := k8s.NewConfMap()
	errConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: controllerConfName}, controlConf)
	//get ingress configs
	controlIngressConf := k8s.NewConfMap()
	errIngressConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: ingressConfigs}, controlIngressConf)
	//get openshift configs
	controlOpenshiftConf := k8s.NewConfMap()
	errOpenshiftConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: openShiftConfigs}, controlOpenshiftConf)
	//get docker registry configs
	dockerRegistryConf := k8s.NewConfMap()
	errRegConf := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: dockerRegConfigs}, dockerRegistryConf)

	confErrs := []error{errConf, errIngressConf, errRegConf, errOpenshiftConf}
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

	registryTypeStr := dockerRegistryConf.Data[registryTypeConst]
	if !registry.IsRegistryType(registryTypeStr) {
		log.Error(err, "Invalid registry type", "registry-type", registryTypeStr)
		// Registry type is invalid, user should update this with valid type.
		// Return and requeue
		return reconcile.Result{}, err
	}
	registryType := registry.Type(dockerRegistryConf.Data[registryTypeConst])
	repositoryName := dockerRegistryConf.Data[repositoryNameConst]
	operatorMode := controlConfigData[operatorModeConst]

	// log controller configurations
	reqLogger.Info(
		"Controller configurations",
		"mgw_toolkit_image", kaniko.DocFileProp.ToolkitImage,
		"mgw_runtime_image", kaniko.DocFileProp.RuntimeImage,
		"kaniko_image", controlConfigData[kanikoImgConst],
		"registry_type", registryType,
		"repository_name", repositoryName,
		"user_nameSpace", userNamespace,
		"operator_mode", operatorMode,
	)

	// handles policy.yaml
	//If there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies
	policyEr := policyHandler(r, operatorOwner, userNamespace)
	if policyEr != nil {
		log.Error(policyEr, "Error in default policy map creation")
	}

	swaggerCmNames := instance.Spec.Definition.SwaggerConfigmapNames
	for _, swaggerCmName := range swaggerCmNames {
		// Check if the configmaps mentioned in the crd object exist
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

		// update owner reference to the swagger configmap and update it
		_ = k8s.UpdateOwner(&r.client, owner, swaggerConfMap)

		// fetch swagger data from configmap and loads as open api swagger v3
		swaggerFileName, swaggerData, _ := config.GetMapKeyValue(swaggerConfMap.Data)
		swaggerFileName = str.GetRandFileName(swaggerFileName) // randomize file name to make it unique
		swaggerOriginal, err := swagger.GetSwaggerV3(&swaggerData)
		if err != nil {
			return reconcile.Result{}, nil
		}

		// Set deployment mode: sidecar/private-jet
		var mode string
		if len(swaggerCmNames) == 1 {
			// override 'instance.Spec.Mode' if there is only one swagger
			mode = swagger.GetMode(swaggerOriginal)
		} else {
			// override mode in swaggers if there are multiple swaggers
			if instance.Spec.Mode != "" {
				mode = instance.Spec.Mode.String()
				log.Info("Set deployment mode in multi swagger mode given in API crd", "mode", mode)
			} else {
				// if no defined in swagger or CRD mode set default
				mode = privateJet
				log.Info("Set deployment mode in multi swagger mode with default mode", "mode", mode)
			}
		}

		apiVersion = swaggerOriginal.Info.Version
		endpointNames := swagger.HandleMgwEndpoints(&r.client, swaggerOriginal, mode, userNamespace)
		apiBasePath := swagger.GetApiBasePath(swaggerOriginal)
		apiBasePathMap[apiBasePath] = apiVersion

		// Creating sidecar endpoint deployment
		if mode == sidecar {
			err := volume.AddSidecarContainers(&r.client, userNamespace, &endpointNames)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		reqLogger.Info("getting security instance")
		//check security scheme already exist
		_, secSchemeDefined := swaggerOriginal.Extensions[swagger.SecuritySchemeExtension]

		securityMap, isDefinedSecurity, resourceLevelSec, securityErr := swagger.GetSecurityMap(swaggerOriginal)
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
			swaggerOriginal.Components.Extensions[swagger.SecuritySchemeExtension] = securityDefinition
		}

		formattedSwagger := swagger.PrettyString(swaggerOriginal)
		formattedSwaggerCmName := swaggerCmName + "-mgw"
		//create configmap with modified swagger
		swaggerDataMgw := map[string]string{swaggerFileName: formattedSwagger}
		swaggerConfMapMgw := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNamespace, Name: formattedSwaggerCmName}, &swaggerDataMgw, nil, owner)
		log.Info("Creating swagger configmap for mgw", "name", formattedSwaggerCmName, "namespace", userNamespace)

		mgwSwaggerConfMap := k8s.NewConfMap()
		errGetConf := k8s.Get(&r.client, types.NamespacedName{Namespace: userNamespace, Name: formattedSwaggerCmName}, mgwSwaggerConfMap)
		if errGetConf != nil && errors.IsNotFound(errGetConf) {
			log.Info("swagger-mgw is not found. Hence creating new configmap")
			_ = k8s.Create(&r.client, swaggerConfMapMgw)
		} else if errGetConf == nil {
			if instance.Spec.UpdateTimeStamp != "" {
				log.Info("Updating swagger-mgw since timestamp value is given")
				_ = k8s.Update(&r.client, swaggerConfMapMgw)
			}
		}

		// Default security
		if !isDefinedSecurity && resourceLevelSec == 0 {
			log.Info("Use default security")

			err := security.Default(&r.client, userNamespace, owner)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	// micro-gateway image to be build
	builtImage := strings.ToLower(strings.ReplaceAll(instance.Name, " ", ""))
	builtImageTag := apiVersion
	// if multi swagger mode override image tag
	if len(swaggerCmNames) > 1 {
		if instance.Spec.Version != "" {
			builtImageTag = instance.Spec.Version
		} else {
			// if not defined in the API Crd set default
			builtImageTag = apiCrdDefaultVersion
		}
	}
	if instance.Spec.UpdateTimeStamp != "" {
		builtImageTag = builtImageTag + "-" + instance.Spec.UpdateTimeStamp
	}
	registry.SetRegistry(registryType, repositoryName, builtImage, builtImageTag)

	// check if the image already exists
	imageExist, errImage := isImageExist(r, utils.DockerRegCredSecret, wso2NameSpaceConst)
	if errImage != nil {
		log.Error(errImage, "Error in image finding")
	}
	log.Info("Is MGW runtime image exist in the docker registry?",
		"exists", strconv.FormatBool(imageExist),
		"registry_type", registryType, "repository_name", repositoryName,
	)

	// handling analytics
	log.Info("Handling analytics")
	if err := analytics.Handle(&r.client, userNamespace); err != nil {
		return reconcile.Result{}, err
	}

	// handling interceptors
	log.Info("Handling interceptors")
	if err := interceptors.Handle(&r.client, instance, owner); err != nil {
		log.Error(err, "Error handling interceptors")
		return reconcile.Result{}, err
	}

	// handling Kaniko docker file
	log.Info("rendering the docker file for Kaniko job and adding volumes to the Kaniko job")
	if err := kaniko.HandleDockerFile(&r.client, userNamespace, instance.Name, owner); err != nil {
		log.Error(err, "Error rendering the docker file for Kaniko job and adding volumes to the Kaniko job")
		return reconcile.Result{}, err
	}

	// sets the MGW configs from APIM configmap
	log.Info("sets the MGW configs from APIM configmap")
	if err := mgw.SetApimConfigs(&r.client); err != nil {
		return reconcile.Result{}, err
	}

	// Retrieving configmap related to micro-gateway configuration mustache/template
	confTemplate := k8s.NewConfMap()
	confErr := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: mgwConfMustache}, confTemplate)
	if confErr != nil {
		log.Error(err, "error in retrieving the config map ")
	}
	//retrieve micro-gw-conf from the configmap
	confTemp := confTemplate.Data[mgwConfGoTmpl]

	//generate mgw conf from the template
	mgwConftmpl, err := template.New("").Parse(confTemp)
	if err != nil {
		log.Error(err, "error in rendering mgw conf with template")
	}
	builder := &strings.Builder{}
	err = mgwConftmpl.Execute(builder, mgw.Configs)
	if err != nil {
		log.Error(err, "error in generating Dockerfile")
	}
	//creating k8s secret from the rendered mgw-conf file
	output := builder.String()

	// create mgwSecret in the k8s cluster
	mgwNsName := types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name + "-" + mgwConfSecretConst}
	mgwData := map[string][]byte{mgwConfConst: []byte(output)}
	mgwSecret := k8s.NewSecretWith(mgwNsName, &mgwData, nil, owner)
	_ = k8s.Apply(&r.client, mgwSecret)

	generateK8sArtifactsForMgw := controlConfigData[generatekubernbetesartifactsformgw]
	genArtifacts, errGenArtifacts := strconv.ParseBool(generateK8sArtifactsForMgw)
	if errGenArtifacts != nil {
		log.Error(errGenArtifacts, "error reading value for generate k8s artifacts")
	}
	getResourceReqCPU := controlConfigData[resourceRequestCPU]
	getResourceReqMemory := controlConfigData[resourceRequestMemory]
	getResourceLimitCPU := controlConfigData[resourceLimitCPU]
	getResourceLimitMemory := controlConfigData[resourceLimitMemory]

	analyticsEnabledBool, _ := strconv.ParseBool(mgw.Configs.AnalyticsEnabled)
	dep := createMgwDeployment(instance, controlConf, analyticsEnabledBool, r, userNamespace, *owner,
		getResourceReqCPU, getResourceReqMemory, getResourceLimitCPU, getResourceLimitMemory, *volume.ContainerList,
		mgw.Configs.HttpPort, mgw.Configs.HttpsPort)
	depFound := &appsv1.Deployment{}
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)

	svc := createMgwLBService(r, instance, userNamespace, *owner, mgw.Configs.HttpPort, mgw.Configs.HttpsPort, operatorMode)
	svcFound := &corev1.Service{}
	svcErr := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, svcFound)

	getMaxRep := controlConfigData[hpaMaxReplicas]
	intValueRep, err := strconv.ParseInt(getMaxRep, 10, 32)
	if err != nil {
		log.Error(err, "error getting max replicas")
	}
	maxReplicas := int32(intValueRep)
	GetAvgUtilCPU := controlConfigData[hpaTargetAverageUtilizationCPU]
	intValueUtilCPU, err := strconv.ParseInt(GetAvgUtilCPU, 10, 32)
	if err != nil {
		log.Error(err, "error getting hpa target average utilization for CPU")
	}
	targetAvgUtilizationCPU := int32(intValueUtilCPU)
	minReplicas := int32(instance.Spec.Replicas)

	err = k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: kanikoArgsConfigs}, kanikoArgs)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No kaniko-arguments config map is available in wso2-system namespace")
	}

	// copy configs and secrets to user's namespace
	if err = copyConfigVolumes(r, userNamespace); err != nil {
		log.Error(err, "Error coping registry specific configs to user's namespace", "user's namespace", userNamespace)
	}

	// Add Kaniko job specific volumes
	volume.AddDefaultKanikoVolumes(instance.Name, swaggerCmNames)

	if instance.Spec.UpdateTimeStamp != "" {
		//Schedule Kaniko pod
		reqLogger.Info("Updating the API", "API.Name", instance.Name, "API.Namespace", instance.Namespace)
		job := scheduleKanikoJob(instance, controlConf, *volume.JobVolumeMount, *volume.JobVolume, instance.Spec.UpdateTimeStamp, owner)
		if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		kubeJob := &batchv1.Job{}
		jobErr := r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, kubeJob)
		// if Job is not available
		if jobErr != nil && errors.IsNotFound(jobErr) {
			reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			jobErr = r.client.Create(context.TODO(), job)
			if jobErr != nil {
				return reconcile.Result{}, jobErr
			}
		} else if jobErr != nil {
			return reconcile.Result{}, jobErr
		}

		// if kaniko job is succeeded, edit the deployment
		if kubeJob.Status.Succeeded > 0 {
			if genArtifacts {
				reqLogger.Info("Job completed successfully", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
				if deperr != nil && errors.IsNotFound(deperr) {
					reqLogger.Info("Creating a new Dep", "Dep.Namespace", dep.Namespace, "Dep.Name", dep.Name)
					deperr = r.client.Create(context.TODO(), dep)
					if deperr != nil {
						return reconcile.Result{}, deperr
					}
					// deployment created successfully - go to create service
				} else if deperr != nil {
					return reconcile.Result{}, deperr
				}
				reqLogger.Info("Updating the found deployment", "Dep.Namespace", dep.Namespace, "Dep.Name", dep.Name)
				updateEr := r.client.Update(context.TODO(), dep)
				if updateEr != nil {
					log.Error(updateEr, "Error in updating deployment")
					return reconcile.Result{}, updateEr
				}
				reqLogger.Info("Skip reconcile: Deployment updated", "Dep.Name", depFound.Name)

				if svcErr != nil && errors.IsNotFound(svcErr) {
					reqLogger.Info("Creating a new Service", "SVC.Namespace", svc.Namespace, "SVC.Name", svc.Name)
					svcErr = r.client.Create(context.TODO(), svc)
					if svcErr != nil {
						return reconcile.Result{}, svcErr
					}

				} else if svcErr != nil {
					return reconcile.Result{}, svcErr
				} else {
					// if service already exsits
					reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
						svcFound.Namespace, "SVC.Name", svcFound.Name)
				}

				errGettingHpa := createHorizontalPodAutoscaler(dep, r, owner, minReplicas, maxReplicas, targetAvgUtilizationCPU)
				if errGettingHpa != nil {
					log.Error(errGettingHpa, "Error getting HPA")
					return reconcile.Result{}, errGettingHpa
				}

				reqLogger.Info("Operator mode is set to " + operatorMode)
				if strings.EqualFold(operatorMode, ingressMode) {
					ingErr := createorUpdateMgwIngressResource(r, instance, mgw.Configs.HttpPort, mgw.Configs.HttpsPort,
						apiBasePathMap, controlIngressConf, owner)
					if ingErr != nil {
						return reconcile.Result{}, ingErr
					}
				}
				if strings.EqualFold(operatorMode, routeMode) {
					rutErr := createorUpdateMgwRouteResource(r, instance, mgw.Configs.HttpPort,
						mgw.Configs.HttpsPort, apiBasePathMap, controlOpenshiftConf, owner)
					if rutErr != nil {
						return reconcile.Result{}, rutErr
					}
				}

				return reconcile.Result{}, nil
			} else {
				log.Info("skip updating kubernetes artifacts")
				return reconcile.Result{}, nil
			}
		} else {
			reqLogger.Info("Job is still not completed.", "Job.Status", job.Status)
			return reconcile.Result{Requeue: true}, nil
		}

	} else if imageExist && !instance.Spec.Override {
		log.Info("Image already exist, hence skipping the kaniko job")

		if genArtifacts {
			log.Info("generating kubernetes artifacts")
			if deperr != nil && errors.IsNotFound(deperr) {
				log.Info("Creating a new Dep", "Dep.Namespace", dep.Namespace, "Dep.Name", dep.Name)
				deperr = r.client.Create(context.TODO(), dep)
				if deperr != nil {
					return reconcile.Result{}, deperr
				}
				// deployment created successfully - go to create service
			} else if deperr != nil {
				return reconcile.Result{}, deperr
			}

			if svcErr != nil && errors.IsNotFound(svcErr) {
				log.Info("Creating a new Service", "SVC.Namespace", svc.Namespace, "SVC.Name", svc.Name)
				svcErr = r.client.Create(context.TODO(), svc)
				if svcErr != nil {
					return reconcile.Result{}, svcErr
				}

			} else if svcErr != nil {
				return reconcile.Result{}, svcErr
			} else {
				// if service already exsits
				reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
					svcFound.Namespace, "SVC.Name", svcFound.Name)
			}

			errGettingHpa := createHorizontalPodAutoscaler(dep, r, owner, minReplicas, maxReplicas, targetAvgUtilizationCPU)
			if errGettingHpa != nil {
				log.Error(errGettingHpa, "Error getting HPA")
				return reconcile.Result{}, errGettingHpa
			}

			reqLogger.Info("Operator mode is set to " + operatorMode)
			if strings.EqualFold(operatorMode, ingressMode) {
				ingErr := createorUpdateMgwIngressResource(r, instance, mgw.Configs.HttpPort, mgw.Configs.HttpsPort,
					apiBasePathMap, controlIngressConf, owner)
				if ingErr != nil {
					return reconcile.Result{}, ingErr
				}
			}
			if strings.EqualFold(operatorMode, routeMode) {
				rutErr := createorUpdateMgwRouteResource(r, instance, mgw.Configs.HttpPort,
					mgw.Configs.HttpsPort, apiBasePathMap, controlOpenshiftConf, owner)
				if rutErr != nil {
					return reconcile.Result{}, rutErr
				}
			}
			return reconcile.Result{}, nil

		} else {
			log.Info("skip generating kubernetes artifacts")
		}

		return reconcile.Result{}, nil
	} else {
		//Schedule Kaniko pod
		job := scheduleKanikoJob(instance, controlConf, *volume.JobVolumeMount, *volume.JobVolume, instance.Spec.UpdateTimeStamp, owner)
		if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		kubeJob := &batchv1.Job{}
		jobErr := r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, kubeJob)
		// if Job is not available
		if jobErr != nil && errors.IsNotFound(jobErr) {
			reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			jobErr = r.client.Create(context.TODO(), job)
			if jobErr != nil {
				return reconcile.Result{}, jobErr
			}
		} else if jobErr != nil {
			return reconcile.Result{}, jobErr
		}

		if kubeJob.Status.Succeeded > 0 {
			reqLogger.Info("Job completed successfully", "Job.Namespace", job.Namespace, "Job.Name", job.Name)

			if genArtifacts {
				if deperr != nil && errors.IsNotFound(deperr) {
					reqLogger.Info("Creating a new Deployment", "Dep.Namespace", dep.Namespace, "Dep.Name", dep.Name)
					deperr = r.client.Create(context.TODO(), dep)
					if deperr != nil {
						return reconcile.Result{}, deperr
					}
					// deployment created successfully - go to create service
				} else if deperr != nil {
					return reconcile.Result{}, deperr
				}
				if svcErr != nil && errors.IsNotFound(svcErr) {
					reqLogger.Info("Creating a new Service", "SVC.Namespace", svc.Namespace, "SVC.Name", svc.Name)
					svcErr = r.client.Create(context.TODO(), svc)
					if svcErr != nil {
						return reconcile.Result{}, svcErr
					}

				} else if svcErr != nil {
					return reconcile.Result{}, svcErr
				} else {
					// if service already exsits
					reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
						svcFound.Namespace, "SVC.Name", svcFound.Name)
				}

				errGettingHpa := createHorizontalPodAutoscaler(dep, r, owner, minReplicas, maxReplicas, targetAvgUtilizationCPU)
				if errGettingHpa != nil {
					log.Error(errGettingHpa, "Error getting HPA")
					return reconcile.Result{}, errGettingHpa
				}

				reqLogger.Info("Operator mode is set to " + operatorMode)
				if strings.EqualFold(operatorMode, ingressMode) {
					ingErr := createorUpdateMgwIngressResource(r, instance, mgw.Configs.HttpPort,
						mgw.Configs.HttpsPort, apiBasePathMap, controlIngressConf, owner)
					if ingErr != nil {
						return reconcile.Result{}, ingErr
					}
				}
				if strings.EqualFold(operatorMode, routeMode) {
					rutErr := createorUpdateMgwRouteResource(r, instance, mgw.Configs.HttpPort, mgw.Configs.HttpsPort, apiBasePathMap, controlOpenshiftConf, owner)
					if rutErr != nil {
						return reconcile.Result{}, rutErr
					}
				}

				return reconcile.Result{}, nil
			} else {
				log.Info("Skip generating kubernetes artifacts")
				return reconcile.Result{}, nil
			}
		} else {
			reqLogger.Info("Job is still not completed.", "Job.Status", job.Status)
			return reconcile.Result{}, deperr
		}
	}
}

// copyConfigVolumes copy the configured secrets and config maps to user's namespace
func copyConfigVolumes(r *ReconcileAPI, namespace string) error {
	config := registry.GetConfig()
	for _, volume := range config.Volumes {
		var fromObj runtime.Object
		var name string

		if volume.Secret != nil {
			name = volume.Secret.SecretName
			fromObj = k8s.NewSecret()
		}
		if volume.ConfigMap != nil {
			name = volume.ConfigMap.Name
			fromObj = k8s.NewConfMap()
		}

		fromNsName := types.NamespacedName{Namespace: wso2NameSpaceConst, Name: name}
		if err := k8s.Get(&r.client, fromNsName, fromObj); err != nil {
			return err
		}

		toObj := fromObj
		toObj.(metav1.Object).SetNamespace(namespace)
		toObjMeta := toObj.(metav1.Object)
		toObjMeta.SetResourceVersion("")
		//newObjMeta := metav1.ObjectMeta{Namespace:namespace, Name:name}

		if err := k8s.Apply(&r.client, toObj); err != nil {
			return err
		}
	}

	return nil
}

func createHorizontalPodAutoscaler(dep *appsv1.Deployment, r *ReconcileAPI, owner *[]metav1.OwnerReference,
	minReplicas int32, maxReplicas int32, targetAverageUtilizationCPU int32) error {

	targetResource := v2beta1.CrossVersionObjectReference{
		Kind:       dep.Kind,
		Name:       dep.Name,
		APIVersion: dep.APIVersion,
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
			OwnerReferences: *owner,
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
		log.Info("Creating HPA for deployment " + dep.Name)
		errHpaCreating := r.client.Create(context.TODO(), hpa)
		if errHpaCreating != nil {
			return errHpaCreating
		}
		return nil
	} else if hpaErr != nil {
		return hpaErr
	} else {
		log.Info("HPA for deployment " + dep.Name + " is already exist")
	}
	return nil
}

// generate relevant MGW deployment/services for the given API definition
func createMgwDeployment(cr *wso2v1alpha1.API, conf *corev1.ConfigMap, analyticsEnabled bool,
	r *ReconcileAPI, nameSpace string, owner []metav1.OwnerReference, resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string, containerList []corev1.Container, httpPortVal int32,
	httpsPortVal int32) *appsv1.Deployment {
	regConfig := registry.GetConfig()
	labels := map[string]string{
		"app": cr.Name,
	}
	controlConfigData := conf.Data
	liveDelay, _ := strconv.ParseInt(controlConfigData[livenessProbeInitialDelaySeconds], 10, 32)
	livePeriod, _ := strconv.ParseInt(controlConfigData[livenessProbePeriodSeconds], 10, 32)
	readDelay, _ := strconv.ParseInt(controlConfigData[readinessProbeInitialDelaySeconds], 10, 32)
	readPeriod, _ := strconv.ParseInt(controlConfigData[readinessProbePeriodSeconds], 10, 32)
	reps := int32(cr.Spec.Replicas)
	var deployVolumeMount []corev1.VolumeMount
	var deployVolume []corev1.Volume
	if analyticsEnabled {
		deployVolumeMountTemp, deployVolumeTemp, err := getAnalyticsPVClaim(r, deployVolumeMount, deployVolume)
		if err != nil {
			log.Error(err, "Analytics volume mounting error")
		} else {
			deployVolumeMount = deployVolumeMountTemp
			deployVolume = deployVolumeTemp
		}
	}
	req := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceReqCPU),
		corev1.ResourceMemory: resource.MustParse(resourceReqMemory),
	}
	lim := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(resourceLimitCPU),
		corev1.ResourceMemory: resource.MustParse(resourceLimitMemory),
	}
	apiContainer := corev1.Container{
		Name:            "mgw" + cr.Name,
		Image:           regConfig.ImagePath,
		ImagePullPolicy: "Always",
		Resources: corev1.ResourceRequirements{
			Requests: req,
			Limits:   lim,
		},
		VolumeMounts: deployVolumeMount,
		Env:          regConfig.Env,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: httpPortVal,
			},
			{
				ContainerPort: httpsPortVal,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.IntOrString{Type: intstr.Int, IntVal: httpsPortVal},
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: int32(readDelay),
			PeriodSeconds:       int32(readPeriod),
			TimeoutSeconds:      1,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.IntOrString{Type: intstr.Int, IntVal: httpsPortVal},
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: int32(liveDelay),
			PeriodSeconds:       int32(livePeriod),
			TimeoutSeconds:      1,
		},
	}

	containerList = append(containerList, apiContainer)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersionKey,
			Kind:       deploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.Name,
			Namespace:       nameSpace,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &reps,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:       containerList,
					Volumes:          deployVolume,
					ImagePullSecrets: regConfig.ImagePullSecrets,
				},
			},
		},
	}
}

func policyHandler(r *ReconcileAPI, operatorOwner *[]metav1.OwnerReference, userNameSpace string) error {
	//Check if policy configmap is available
	foundmapc := k8s.NewConfMap()
	err := k8s.Get(&r.client, types.NamespacedName{Name: policyConfigmap, Namespace: userNameSpace}, foundmapc)

	if err != nil && errors.IsNotFound(err) {
		//create new map with default policies in user namespace if a map is not found
		log.Info("Creating a config map with default policies", "Namespace", userNameSpace, "Name", policyConfigmap)

		defaultval := ratelimiting.CreateDefault()
		policyDataMap := map[string]string{policyFileConst: defaultval}
		policyConfMap := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNameSpace, Name: policyConfigmap}, &policyDataMap, nil, operatorOwner)

		err = r.client.Create(context.TODO(), policyConfMap)
		if err != nil {
			log.Error(err, "error ")
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// isImageExist checks if the image exists in the given registry using the secret in the user-namespace
func isImageExist(r *ReconcileAPI, secretName string, namespace string) (bool, error) {
	var registryUrl string
	var username string
	var password string

	type Auth struct {
		Auths map[string]struct {
			Auth     string `json:"auth"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auths"`
	}

	// checks if the secret is available
	log.Info("Getting Docker credentials secret")
	dockerConfigSecret := k8s.NewSecret()
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, dockerConfigSecret)
	if err == nil && errors.IsNotFound(err) {
		log.Info("Docker credentials secret is not found", "secret-name", secretName, "namespace", namespace)
	} else if err != nil {
		authsJsonString := dockerConfigSecret.Data[utils.DockerConfigKeyConst]
		auths := Auth{}
		err := json.Unmarshal([]byte(authsJsonString), &auths)
		if err != nil {
			log.Info("Error unmarshal data of docker credential auth")
		}

		for regUrl, credential := range auths.Auths {
			registryUrl = str.RemoveVersionTag(regUrl)
			if !strings.HasPrefix(registryUrl, "https://") {
				registryUrl = "https://" + registryUrl
			}
			username = credential.Username
			password = credential.Password

			break
		}
	}

	return registry.IsImageExists(utils.RegAuth{RegistryUrl: registryUrl, Username: username, Password: password}, log)
}

//Schedule Kaniko Job to generate micro-gw image
func scheduleKanikoJob(cr *wso2v1alpha1.API, conf *corev1.ConfigMap, jobVolumeMount []corev1.VolumeMount,
	jobVolume []corev1.Volume, timeStamp string, owner *[]metav1.OwnerReference) *batchv1.Job {
	roolValue := int64(0)
	regConfig := registry.GetConfig()
	kanikoJobName := cr.Name + "-kaniko"
	if timeStamp != "" {
		kanikoJobName = kanikoJobName + "-" + timeStamp
	}
	controlConfigData := conf.Data
	kanikoImg := controlConfigData[kanikoImgConst]
	//read kaniko arguments and split them as they are read as a single string
	kanikoArguments := strings.Split(kanikoArgs.Data[kanikoArguments], "\n")
	args := append([]string{
		"--dockerfile=/usr/wso2/dockerfile/Dockerfile",
		"--context=/usr/wso2/",
		"--destination=" + regConfig.ImagePath,
	}, regConfig.Args...)
	//if kanikoarguments are provided
	if kanikoArguments != nil {
		args = append(args, kanikoArguments...)
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kanikoJobName,
			Namespace:       cr.Namespace,
			OwnerReferences: *owner,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name + "-job",
					Namespace: cr.Namespace,
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         cr.Name + "gen-container",
							Image:        kanikoImg,
							VolumeMounts: jobVolumeMount,
							Args:         args,
							Env:          regConfig.Env,
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &roolValue,
					},
					RestartPolicy: "Never",
					Volumes:       jobVolume,
				},
			},
		},
	}
}

//Creating a LB balancer service to expose mgw
func createMgwLBService(r *ReconcileAPI, cr *wso2v1alpha1.API, nameSpace string, owner []metav1.OwnerReference, httpPortVal int32,
	httpsPortVal int32, operatorMode string) *corev1.Service {
	var serviceType corev1.ServiceType
	serviceType = corev1.ServiceTypeLoadBalancer

	if strings.EqualFold(operatorMode, ingressMode) || strings.EqualFold(operatorMode, clusterIPMode) ||
		strings.EqualFold(operatorMode, routeMode) {
		serviceType = corev1.ServiceTypeClusterIP
	}

	labels := map[string]string{
		"app": cr.Name,
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.Name,
			Namespace:       nameSpace,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{{
				Name:       httpsConst + "-" + portConst,
				Port:       httpsPortVal,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: httpsPortVal},
			}, {
				Name:       httpConst + "-" + portConst,
				Port:       httpPortVal,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: httpPortVal},
			}},
			Selector: labels,
		},
	}

	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

// Creating an Ingress resource to expose mgw
// Supports for multiple apiBasePaths when there are multiple swaggers for one API CRD
func createorUpdateMgwIngressResource(r *ReconcileAPI, cr *wso2v1alpha1.API, httpPortVal int32, httpsPortVal int32,
	apiBasePathMap map[string]string, controllerConfig *corev1.ConfigMap, owner *[]metav1.OwnerReference) error {

	controlConfigData := controllerConfig.Data
	transportMode := controlConfigData[ingressTransportMode]
	ingressHostName := controlConfigData[ingressHostName]
	tlsSecretName := controlConfigData[tlsSecretName]
	ingressNamePrefix := controlConfigData[ingressResourceName]
	ingressName := ingressNamePrefix + "-" + cr.Name
	namespace := cr.Namespace
	apiServiceName := cr.Name

	var hostArray []string
	hostArray = append(hostArray, ingressHostName)
	log.Info(fmt.Sprintf("Creating ingress resource with name: %v", ingressName))
	log.WithValues("Ingress metadata. Transport mode", transportMode, "Ingress name", ingressName,
		"Ingress hostname ", ingressHostName)
	annotationMap := k8s.NewConfMap()
	err := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: ingressConfigs}, annotationMap)
	var port int32

	if httpConst == transportMode {
		port = httpPortVal
	} else {
		port = httpsPortVal
	}

	annotationConfigData := annotationMap.Data
	annotationsList := annotationConfigData[ingressProperties]
	var ingressAnnotationMap map[string]string
	ingressAnnotationMap = make(map[string]string)

	splitArray := strings.Split(annotationsList, "\n")
	for _, element := range splitArray {
		if element != "" && strings.ContainsAny(element, ":") {
			splitValues := strings.Split(element, ":")
			ingressAnnotationMap[strings.TrimSpace(splitValues[0])] = strings.TrimSpace(splitValues[1])
		}
	}

	log.Info("Creating ingress resource with the following Base Paths")

	// add multiple api base paths
	var httpIngressPaths []v1beta1.HTTPIngressPath
	for basePath := range apiBasePathMap {

		apiBasePath := basePath
		// if the base path contains /petstore/{version}, then it is converted to /petstore/1.0.0
		if strings.Contains(basePath, versionField) {
			apiBasePath = strings.Replace(basePath, versionField, apiBasePathMap[basePath], -1)
		}

		log.Info(fmt.Sprintf("Adding the base path: %v to ingress resource", apiBasePath))
		httpIngressPaths = append(httpIngressPaths, v1beta1.HTTPIngressPath{
			Path: apiBasePath,
			Backend: v1beta1.IngressBackend{
				ServiceName: apiServiceName,
				ServicePort: intstr.IntOrString{IntVal: port},
			},
		})

	}

	ingressResource := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       namespace, // goes into backend full name
			Name:            ingressName,
			Annotations:     ingressAnnotationMap,
			OwnerReferences: *owner,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressHostName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: httpIngressPaths,
						},
					},
				},
			},
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      hostArray,
					SecretName: tlsSecretName,
				},
			},
		},
	}

	ingress := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingressName, Namespace: namespace}, ingress)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Ingress resource not found with name " + ingressName + ".Hence creating a new Ingress resource")
		err = r.client.Create(context.TODO(), ingressResource)
		return err
	} else {
		log.Info("Ingress resource found with name " + ingressName + ".Hence updating the existing Ingress resource")
		err = r.client.Update(context.TODO(), ingressResource)
		return err
	}
	return err
}

// Creating a Route resource to expose microgateway
// Supports for multiple apiBasePaths when there are multiple swaggers for one API CRD
func createorUpdateMgwRouteResource(r *ReconcileAPI, cr *wso2v1alpha1.API, httpPortVal int32, httpsPortVal int32,
	apiBasePathMap map[string]string, controllerConfig *corev1.ConfigMap, owner *[]metav1.OwnerReference) error {

	controlConfigData := controllerConfig.Data
	routePrefix := controlConfigData[routeName]
	routesHostname := controlConfigData[routeHost]
	transportMode := controlConfigData[routeTransportMode]
	tlsTerminationValue := controlConfigData[tlsTermination]

	var tlsTerminationType routv1.TLSTerminationType
	if strings.EqualFold(tlsTerminationValue, edge) {
		tlsTerminationType = routv1.TLSTerminationEdge
	} else if strings.EqualFold(tlsTerminationValue, reencrypt) {
		tlsTerminationType = routv1.TLSTerminationReencrypt
	} else if strings.EqualFold(tlsTerminationValue, passthrough) {
		tlsTerminationType = routv1.TLSTerminationPassthrough
	} else {
		tlsTerminationType = ""
	}

	routeName := routePrefix + "-" + cr.Name
	namespace := cr.Namespace
	apiServiceName := cr.Name

	var hostArray []string
	hostArray = append(hostArray, routesHostname)
	log.Info(fmt.Sprintf("Creating route resource with name: %v", routeName))
	log.WithValues("Route metadata. Transport mode", transportMode, "Route name", routeName,
		"Ingress hostname ", routesHostname)

	annotationMap := k8s.NewConfMap()
	err := k8s.Get(&r.client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: openShiftConfigs}, annotationMap)
	var port int32

	if httpConst == transportMode {
		port = httpPortVal
	} else {
		port = httpsPortVal
	}

	annotationConfigData := annotationMap.Data
	annotationsList := annotationConfigData[routeProperties]
	var routeAnnotationMap map[string]string
	routeAnnotationMap = make(map[string]string)

	splitArray := strings.Split(annotationsList, "\n")
	for _, element := range splitArray {
		if element != "" && strings.ContainsAny(element, ":") {
			splitValues := strings.Split(element, ":")
			routeAnnotationMap[strings.TrimSpace(splitValues[0])] = strings.TrimSpace(splitValues[1])
		}
	}

	log.Info("Creating route resource for API " + cr.Name)

	var routeList []routv1.Route

	for basePath := range apiBasePathMap {

		apiBasePath := basePath
		// if the base path contains /petstore/{version}, then it is converted to /petstore/1.0.0
		if strings.Contains(basePath, versionField) {
			apiBasePath = strings.Replace(basePath, versionField, apiBasePathMap[basePath], -1)
		}

		apiBasePathSuffix := apiBasePath
		apiBasePathSuffix = strings.Replace(apiBasePathSuffix, "/", "-", -1)
		routeNewName := routeName + apiBasePathSuffix

		log.Info(fmt.Sprintf("Creating the route : %v to ingress resource", apiBasePath))

		routeResource := routv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:            routeNewName,
				Namespace:       namespace,
				OwnerReferences: *owner,
				Annotations:     routeAnnotationMap,
			},
			Spec: routv1.RouteSpec{
				Host: routesHostname,
				Path: apiBasePath,
				Port: &routv1.RoutePort{
					TargetPort: intstr.IntOrString{IntVal: port},
				},
				To: routv1.RouteTargetReference{
					Kind: serviceKind,
					Name: apiServiceName,
				},
				TLS: &routv1.TLSConfig{
					Termination: tlsTerminationType,
				},
			},
		}

		routeList = append(routeList, routeResource)
	}

	for _, route := range routeList {

		routeGet := &routv1.Route{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, routeGet)

		if err != nil && errors.IsNotFound(err) {
			log.Info("Route resource not found with name " + route.Name + ".Hence creating a new Route resource")
			errInCreating := r.client.Create(context.TODO(), &route)

			if errInCreating != nil {
				return errInCreating
			}
		} else {
			log.Info("Route resource found with name " + route.Name + ".Hence updating the existing Route resource")
			routeGet.Spec = route.Spec
			errInUpdating := r.client.Update(context.TODO(), routeGet)

			if errInUpdating != nil {
				return errInUpdating
			}
		}

	}

	return nil
}

//Mounts an emptydir volume to be used when analytics is enabled
func getAnalyticsPVClaim(r *ReconcileAPI, deployVolumeMount []corev1.VolumeMount, deployVolume []corev1.Volume) ([]corev1.VolumeMount, []corev1.Volume, error) {

	deployVolumeMount = []corev1.VolumeMount{
		{
			Name:      analyticsVolumeName,
			MountPath: analyticsVolumeLocation,
			ReadOnly:  false,
		},
	}
	deployVolume = []corev1.Volume{
		{
			Name: analyticsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	return deployVolumeMount, deployVolume, nil
}

//gets the details of the operator for owner reference
func getOperatorOwner(r *ReconcileAPI) (*[]metav1.OwnerReference, error) {
	depFound := &appsv1.Deployment{}
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: "api-operator", Namespace: wso2NameSpaceConst}, depFound)
	if deperr != nil {
		noOwner := []metav1.OwnerReference{}
		return &noOwner, deperr
	}

	return k8s.NewOwnerRef(depFound.TypeMeta, depFound.ObjectMeta), nil
}

func deleteCompletedJobs(namespace string) error {
	log.Info("Deleting completed kaniko job")
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Can't load in cluster config: %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Can't get client set: %v", err)
		return err
	}

	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}
	//get list of exsisting jobs
	getListOfJobs, errGetJobs := clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if len(getListOfJobs.Items) != 0 {
		for _, kanikoJob := range getListOfJobs.Items {
			if kanikoJob.Status.Succeeded > 0 {
				log.Info("Job "+kanikoJob.Name+" completed successfully", "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				log.Info("Deleting job "+kanikoJob.Name, "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				//deleting completed jobs
				errDelete := clientset.BatchV1().Jobs(kanikoJob.Namespace).Delete(kanikoJob.Name, &deleteOptions)
				if errDelete != nil {
					log.Error(errDelete, "error while deleting "+kanikoJob.Name+" job")
					return errDelete
				} else {
					log.Info("successfully deleted job "+kanikoJob.Name, "Job.Namespace", kanikoJob.Namespace, "Job.Name", kanikoJob.Name)
				}
			}
		}
	} else if errGetJobs != nil {
		log.Error(errGetJobs, "error retrieving jobs")
		return err
	}
	return nil
}
