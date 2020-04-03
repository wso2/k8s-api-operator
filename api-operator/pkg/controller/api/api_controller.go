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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/golang/glog"
	"github.com/sirupsen/logrus"
	v1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/serving/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/url"
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

	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"gopkg.in/yaml.v2"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/ratelimiting"
)

var log = logf.Log.WithName("controller_api")

//XMGWProductionEndpoints represents the structure of endpoint
type XMGWProductionEndpoints struct {
	Urls []string `yaml:"urls" json:"urls"`
}

//This struct use to import multiple certificates to trsutstore
type DockerfileArtifacts struct {
	CertFound             bool
	Password              string
	Certs                 map[string]string
	BaseImage             string
	RuntimeImage          string
	InterceptorsFound     bool
	JavaInterceptorsFound bool
}

//These structs used to build the security schema in json
type path struct {
	Security []map[string][]string `json:"security"`
}
type securitySchemeStruct struct {
	SecurityType string             `json:"type"`
	Scheme       string             `json:"scheme,omitempty"`
	Flows        *authorizationCode `json:"flows,omitempty"`
}

type MGWConf struct {
	KeystorePath                   string
	KeystorePassword               string
	TruststorePath                 string
	TruststorePassword             string
	KeymanagerServerurl            string
	KeymanagerUsername             string
	KeymanagerPassword             string
	JwtConfigs                     []SecurityTypeJWT
	EnabledGlobalTMEventPublishing string
	JmsConnectionProvider          string
	ThrottleEndpoint               string
	EnableRealtimeMessageRetrieval string
	EnableRequestValidation        string
	EnableResponseValidation       string
	LogLevel                       string
	HttpPort                       string
	HttpsPort                      string
	BasicUsername                  string
	BasicPassword                  string
	AnalyticsEnabled               string
	AnalyticsUsername              string
	AnalyticsPassword              string
	UploadingTimeSpanInMillis      string
	RotatingPeriod                 string
	UploadFiles                    string
	VerifyHostname                 string
	Hostname                       string
	Port                           string
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

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
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

	var containerList []corev1.Container      // containerList represents the list of containers for micro-gateway deployment
	var deployedSidecarEndpointNames []string // container already added sidecar endpoint names
	var jobVolume []corev1.Volume             // Volumes for Kaniko Job
	var jobVolumeMount []corev1.VolumeMount   // Volume mounts for Kaniko Job
	var apiBasePaths []string                 // API base paths
	var apiVersion string                     // API version - for the tag of final MGW docker image

	var alias string
	var existCert = false             // keep to track the existence of certificates
	var existBalInterceptors = false  // keep to track the existence of interceptors
	var existJavaInterceptors = false // keep to track the existence of java interceptors
	var certName string

	//get multiple jwt issuer details
	jwtConfigs := []SecurityTypeJWT{}
	var certList = make(map[string]string) // to add multiple certs with alias
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

	owner := getOwnerDetails(instance)
	operatorOwner, ownerErr := getOperatorOwner(r)
	if ownerErr != nil {
		reqLogger.Info("Operator was not found in the " + wso2NameSpaceConst + " namespace. No owner will be set for the artifacts")
	}
	userNameSpace := instance.Namespace

	//get configurations file for the controller
	controlConf, errConf := getConfigmap(r, controllerConfName, wso2NameSpaceConst)
	//get ingress configs
	controlIngressConf, errIngressConf := getConfigmap(r, ingressConfigs, wso2NameSpaceConst)
	//get openshift configs
	controlOpenshiftConf, errOpenshiftConf := getConfigmap(r, openShiftConfigs, wso2NameSpaceConst)
	//get docker registry configs
	dockerRegistryConf, errRegConf := getConfigmap(r, dockerRegConfigs, wso2NameSpaceConst)

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
	mgwToolkitImg := controlConfigData[mgwToolkitImgConst]
	mgwRuntimeImg := controlConfigData[mgwRuntimeImgConst]
	kanikoImg := controlConfigData[kanikoImgConst]

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

	reqLogger.Info("Controller Configurations", "mgwToolkitImg", mgwToolkitImg, "mgwRuntimeImg", mgwRuntimeImg,
		"kanikoImg", kanikoImg, "registryType", registryType, "repositoryName", repositoryName,
		"userNameSpace", userNameSpace, "operatorMode", operatorMode)

	//Handles policy.yaml.
	//If there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies
	policyEr := policyHandler(r, operatorOwner, userNameSpace)
	if policyEr != nil {
		log.Error(policyEr, "Error in default policy map creation")
	}

	// make volumes empty
	jobVolumeMount, jobVolume = []corev1.VolumeMount{}, []corev1.Volume{}

	// Check if the configmaps mentioned in the crd object exist
	swaggerCmNames := instance.Spec.Definition.SwaggerConfigmapNames
	for _, swaggerCmName := range swaggerCmNames {
		apiConfigMap, err := getConfigmap(r, swaggerCmName, userNameSpace)
		if err != nil {
			if errors.IsNotFound(err) {
				// Swagger configmap is not found, could have been deleted after reconcile request.
				// Return and requeue
				log.Error(err, "Swagger configmap is not found")
				return reconcile.Result{}, err
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}

		//add owner reference to the swagger configmap and update it
		errUpdateApiConf := updateConfMapWithOwner(r, owner, apiConfigMap)
		if errUpdateApiConf != nil {
			log.Error(errUpdateApiConf, "error in updating swagger config map with owner reference")
		}

		//Fetch swagger data from configmap, reads and modifies swagger
		swaggerDataMap := apiConfigMap.Data
		swagger, swaggerDataFile, err := mgwSwaggerLoader(swaggerDataMap)
		// randomize file name to make it unique
		swaggerDataFile = getRandFileName(swaggerDataFile)

		// Set deployment mode: sidecar/private-jet
		var mode string
		if len(swaggerCmNames) == 1 {
			// override 'instance.Spec.Mode' if there is only one swagger
			modeExt, isModeDefined := swagger.Extensions[deploymentMode]
			if isModeDefined {
				modeRawStr, _ := modeExt.(json.RawMessage)
				err = json.Unmarshal(modeRawStr, &mode)
				if err != nil {
					log.Info("Error unmarshal data of mode")
				}
			} else {
				log.Info("Deployment mode is not set in the swagger.")
			}
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

		// Set the apiVersion of the swagger if there is only one swagger
		if len(swaggerCmNames) == 1 {
			apiVersion = swagger.Info.Version
		}

		endpointNames, newSwagger, apiBasePath := mgwSwaggerHandler(r, swagger, mode, userNameSpace)
		apiBasePaths = append(apiBasePaths, apiBasePath)
		for endpointNameL, _ := range endpointNames {
			log.Info("Endpoint name " + endpointNameL)
		}

		//Creating sidecar endpoint deployment
		if mode == sidecar {
			for endpointName, _ := range endpointNames {
				// deploy sidecar only if endpoint name is not empty and not already deployed
				if endpointName != "" && !isStringArrayContains(deployedSidecarEndpointNames, endpointName) {
					targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
					erCr := r.client.Get(context.TODO(),
						types.NamespacedName{Namespace: userNameSpace, Name: endpointName}, targetEndpointCr)
					if erCr == nil && targetEndpointCr.Spec.Deploy.DockerImage != "" {
						sidecarContainer := corev1.Container{
							Image: targetEndpointCr.Spec.Deploy.DockerImage,
							Name:  targetEndpointCr.Spec.Deploy.Name,
							Ports: []corev1.ContainerPort{{
								ContainerPort: targetEndpointCr.Spec.Port,
							}},
						}
						containerList = append(containerList, sidecarContainer)
						deployedSidecarEndpointNames = append(deployedSidecarEndpointNames, endpointName)
					} else {
						log.Info("Failed to deploy the sidecar endpoint " + endpointName)
						return reconcile.Result{}, erCr
					}
				}
			}
		}

		reqLogger.Info("getting security instance")
		//check security scheme already exist
		_, secSchemeDefined := swagger.Extensions[securitySchemeExtension]

		securityMap, isDefinedSecurity, resourceLevelSec, securityErr := getSecurityDefinedInSwagger(swagger)
		if securityErr != nil {
			return reconcile.Result{}, securityErr
		}

		securityDefinition, existSecurityCerts, updatedCertList, volumemountTemp, volumeTemp, jwtConfArray, errsec := handleSecurity(r, securityMap, userNameSpace, instance, secSchemeDefined, certList, jobVolumeMount, jobVolume)
		if errsec != nil {
			return reconcile.Result{}, errsec
		}
		certList = updatedCertList
		existCert = existSecurityCerts
		jobVolumeMount = volumemountTemp
		jobVolume = volumeTemp
		jwtConfigs = jwtConfArray

		//adding security scheme to swagger
		if len(securityDefinition) > 0 {
			newSwagger.Components.Extensions[securitySchemeExtension] = securityDefinition
		}
		//reformatting swagger
		var prettyJSON bytes.Buffer
		final, err := newSwagger.MarshalJSON()
		if err != nil {
			log.Error(err, "swagger marshal error")
		}
		errIndent := json.Indent(&prettyJSON, final, "", "  ")
		if errIndent != nil {
			log.Error(errIndent, "Error in pretty json")
		}

		formattedSwagger := string(prettyJSON.Bytes())
		formattedSwaggerCmName := swaggerCmName + "-mgw"
		//create configmap with modified swagger
		swaggerConfMap := createConfigMap(formattedSwaggerCmName, swaggerDataFile, formattedSwagger, userNameSpace, owner)
		log.Info("Creating swagger configmap for mgw", "name", formattedSwaggerCmName, "namespace", userNameSpace)

		_, errGetConf := getConfigmap(r, formattedSwaggerCmName, userNameSpace)
		if errGetConf != nil && errors.IsNotFound(errGetConf) {
			log.Info("swagger-mgw is not found. Hence creating new configmap")
			errCrtConf := r.client.Create(context.TODO(), swaggerConfMap)
			if errCrtConf != nil {
				log.Error(errCrtConf, "Error creating mgw swagger configmap")
			}
		} else if errGetConf != nil {
			log.Error(errGetConf, "error getting swagger-mgw")
		} else {
			if instance.Spec.UpdateTimeStamp != "" {
				log.Info("updating swagger-mgw since timestamp value is given")
				updateEr := r.client.Update(context.TODO(), swaggerConfMap)
				if updateEr != nil {
					log.Error(updateEr, "Error in updating configmap with updated swagger definition")
				}
			}
		}

		if isDefinedSecurity == false && resourceLevelSec == 0 {
			log.Info("use default security")
			//var certName string
			defaultSecConf := SecurityTypeJWT{}
			//use default security
			//copy default sec in wso2-system to user namespace
			securityDefault := &wso2v1alpha1.Security{}
			//check default security already exist in user namespace
			errGetSec := r.client.Get(context.TODO(), types.NamespacedName{Name: defaultSecurity, Namespace: userNameSpace}, securityDefault)

			if errGetSec != nil && errors.IsNotFound(errGetSec) {
				log.Info("default security not found in " + userNameSpace + " namespace")
				log.Info("retrieve default-security from " + wso2NameSpaceConst)
				//retrieve default-security from wso2-system namespace
				errSec := r.client.Get(context.TODO(), types.NamespacedName{Name: defaultSecurity, Namespace: wso2NameSpaceConst}, securityDefault)
				if errSec != nil && errors.IsNotFound(errSec) {
					reqLogger.Info("default security instance is not found in " + wso2NameSpaceConst)
					return reconcile.Result{}, errSec
				} else if errSec != nil {
					log.Error(errSec, "error in getting default security from "+wso2NameSpaceConst)
					return reconcile.Result{}, errSec
				}
				var defaultCert = &corev1.Secret{}
				//check default certificate exists in user namespace
				errCertUserns := r.client.Get(context.TODO(), types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: userNameSpace}, defaultCert)
				if errCertUserns != nil && errors.IsNotFound(errCertUserns) {
					log.Info("default certificate is not found in " + userNameSpace + "namespace")
					log.Info("retrieve default certificate from " + wso2NameSpaceConst)
					var defaultCertName string
					var defaultCertvalue []byte
					errc := r.client.Get(context.TODO(), types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: wso2NameSpaceConst}, defaultCert)
					if errc != nil && errors.IsNotFound(errc) {
						reqLogger.Info("defined certificate is not found in " + wso2NameSpaceConst)
						return reconcile.Result{}, errc
					} else if errc != nil {
						log.Error(errc, "error in getting default cert from "+wso2NameSpaceConst)
						return reconcile.Result{}, errc
					}
					//copying default cert as a secret to user namespace
					for cert, value := range defaultCert.Data {
						defaultCertName = cert
						defaultCertvalue = value
					}
					newDefaultSecret := createSecret(securityDefault.Spec.SecurityConfig[0].Certificate, defaultCertName, string(defaultCertvalue), userNameSpace, owner)
					errCreateSec := r.client.Create(context.TODO(), newDefaultSecret)
					if errCreateSec != nil {
						log.Error(errCreateSec, "error creating secret for default security in user namespace")
						return reconcile.Result{}, errCreateSec
					} else {
						//mount certs
						volumemountTemp, volumeTemp := certMoutHandler(r, newDefaultSecret, jobVolumeMount, jobVolume)
						jobVolumeMount = volumemountTemp
						jobVolume = volumeTemp
						alias = newDefaultSecret.Name + certAlias
						existCert = true
						for k := range newDefaultSecret.Data {
							certName = k
						}
						//add cert path and alias as key value pairs
						certList[alias] = certPath + newDefaultSecret.Name + "/" + certName
						defaultSecConf.CertificateAlias = alias
					}
				} else if errCertUserns != nil {
					log.Error(errCertUserns, "error in getting default certificate from "+userNameSpace+"namespace")
					return reconcile.Result{}, errCertUserns
				} else {
					//mount certs
					volumemountTemp, volumeTemp := certMoutHandler(r, defaultCert, jobVolumeMount, jobVolume)
					jobVolumeMount = volumemountTemp
					jobVolume = volumeTemp
					alias = defaultCert.Name + certAlias
					existCert = true
					for k := range defaultCert.Data {
						certName = k
					}
					//add cert path and alias as key value pairs
					certList[alias] = certPath + defaultCert.Name + "/" + certName
					defaultSecConf.CertificateAlias = alias
				}
				//copying default security to user namespace
				log.Info("copying default security to " + userNameSpace)
				newDefaultSecurity := copyDefaultSecurity(securityDefault, userNameSpace, owner)
				errCreateSecurity := r.client.Create(context.TODO(), newDefaultSecurity)
				if errCreateSecurity != nil {
					log.Error(errCreateSecurity, "error creating secret for default security in user namespace")
					return reconcile.Result{}, errCreateSecurity
				}
				log.Info("default security successfully copied to " + userNameSpace + " namespace")
				if newDefaultSecurity.Spec.SecurityConfig[0].Issuer != "" {
					defaultSecConf.Issuer = newDefaultSecurity.Spec.SecurityConfig[0].Issuer
				}
				if newDefaultSecurity.Spec.SecurityConfig[0].Audience != "" {
					defaultSecConf.Audience = newDefaultSecurity.Spec.SecurityConfig[0].Audience
				}
				defaultSecConf.ValidateSubscription = newDefaultSecurity.Spec.SecurityConfig[0].ValidateSubscription
			} else if errGetSec != nil {
				log.Error(errGetSec, "error getting default security from user namespace")
				return reconcile.Result{}, errGetSec
			} else {
				log.Info("default security exists in " + userNameSpace + " namespace")
				//check default cert exist in usernamespace
				var defaultCertUsrNs = &corev1.Secret{}
				errCertUserns := r.client.Get(context.TODO(), types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: userNameSpace}, defaultCertUsrNs)
				if errCertUserns != nil && errors.IsNotFound(errCertUserns) {
					log.Error(errCertUserns, "default certificate is not found in user namespace")
					return reconcile.Result{}, errCertUserns
				} else if errCertUserns != nil {
					log.Error(errCertUserns, "error retrieving default certificate in user namespace")
					return reconcile.Result{}, errCertUserns
				} else {
					//mount certs
					volumemountTemp, volumeTemp := certMoutHandler(r, defaultCertUsrNs, jobVolumeMount, jobVolume)
					jobVolumeMount = volumemountTemp
					jobVolume = volumeTemp
					alias = defaultCertUsrNs.Name + certAlias
					existCert = true
					for k := range defaultCertUsrNs.Data {
						certName = k
					}
					//add cert path and alias as key value pairs
					certList[alias] = certPath + defaultCertUsrNs.Name + "/" + certName
					certificateAlias = alias
					defaultSecConf.CertificateAlias = alias
					defaultSecConf.ValidateSubscription = securityDefault.Spec.SecurityConfig[0].ValidateSubscription
				}
				if securityDefault.Spec.SecurityConfig[0].Issuer != "" {
					defaultSecConf.Issuer = securityDefault.Spec.SecurityConfig[0].Issuer
				}
				if securityDefault.Spec.SecurityConfig[0].Audience != "" {
					defaultSecConf.Audience = securityDefault.Spec.SecurityConfig[0].Audience
				}
			}
			jwtConfigs = append(jwtConfigs, defaultSecConf)
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
	log.Info("image exist? " + strconv.FormatBool(imageExist))

	// gets analytics configuration
	analyticsConf, analyticsEr := getConfigmap(r, analyticsConfName, wso2NameSpaceConst)
	if analyticsEr != nil {
		log.Info("Disabling analytics since the analytics configuration related config map not found.")
		analyticsEnabled = "false"
	} else {
		if analyticsConf.Data[analyticsEnabledConst] == "true" {
			uploadingTimeSpanInMillis = analyticsConf.Data[uploadingTimeSpanInMillisConst]
			rotatingPeriod = analyticsConf.Data[rotatingPeriodConst]
			uploadFiles = analyticsConf.Data[uploadFilesConst]
			hostname = analyticsConf.Data[hostnameConst]
			port = analyticsConf.Data[portConst]
			analyticsSecretName := analyticsConf.Data[analyticsSecretConst]

			// gets the data from analytics secret
			analyticsData, err := getSecretData(r, analyticsSecretName)

			if err == nil && analyticsData != nil && analyticsData[usernameConst] != nil &&
				analyticsData[passwordConst] != nil && analyticsData[certConst] != nil {
				analyticsUsername = string(analyticsData[usernameConst])
				analyticsPassword = string(analyticsData[passwordConst])
				analyticsCertSecretName := string(analyticsData[certConst])

				log.Info("Finding analytics cert secret " + analyticsCertSecretName)
				//Check if this secret exists and append it to volumes
				jobVolumeMountTemp, jobVolumeTemp, fileName, errCert := analyticsVolumeHandler(analyticsCertSecretName,
					r, jobVolumeMount, jobVolume, userNameSpace, operatorOwner)
				if errCert == nil {
					jobVolumeMount = jobVolumeMountTemp
					jobVolume = jobVolumeTemp
					existCert = true
					analyticsEnabled = "true"
					certList[analyticsAlias] = analyticsCertLocation + fileName
				}
			}
		}
	}

	//Handle interceptors if available
	tmpExistBalInterceptors, tmpExistJavaInterceptors, jobVolumeMountTemp, jobVolumeTemp, errBalInterceptor, errJavaInterceptor := interceptorHandler(r, instance, owner, jobVolumeMount, jobVolume, userNameSpace)
	existBalInterceptors = existBalInterceptors || tmpExistBalInterceptors
	existJavaInterceptors = existJavaInterceptors || tmpExistJavaInterceptors
	jobVolumeMount = jobVolumeMountTemp
	jobVolume = jobVolumeTemp
	if errBalInterceptor != nil || errJavaInterceptor != nil {
		return reconcile.Result{}, errBalInterceptor
	}

	//Handles the creation of dockerfile configmap
	dockerfileConfmap, errDocker := dockerfileHandler(r, certList, existCert, controlConfigData, owner, instance, existBalInterceptors, existJavaInterceptors)
	if errDocker != nil {
		log.Error(errDocker, "error in docker configmap handling")
		return reconcile.Result{}, errDocker
	} else {
		log.Info("kaniko job related dockerfile was written into configmap " + dockerfileConfmap.Name)
	}

	//Get data from apim configmap
	apimConfig, apimEr := getConfigmap(r, apimConfName, wso2NameSpaceConst)
	httpPortVal := httpPortValConst
	httpsPortVal := httpsPortValConst
	if apimEr == nil {
		verifyHostname = apimConfig.Data[verifyHostnameConst]
		enabledGlobalTMEventPublishing = apimConfig.Data[enabledGlobalTMEventPublishingConst]
		jmsConnectionProvider = apimConfig.Data[jmsConnectionProviderConst]
		throttleEndpoint = apimConfig.Data[throttleEndpointConst]
		enableRealtimeMessageRetrieval = apimConfig.Data[enableRealtimeMessageRetrievalConst]
		enableRequestValidation = apimConfig.Data[enableRequestValidationConst]
		enableResponseValidation = apimConfig.Data[enableResponseValidationConst]
		logLevel = apimConfig.Data[logLevelConst]
		httpPort = apimConfig.Data[httpPortConst]
		httpsPort = apimConfig.Data[httpsPortConst]
		httpPortVal, err = strconv.Atoi(httpPort)
		if err != nil {
			log.Error(err, "Valid http port was not provided. Default port will be used")
			httpPortVal = httpPortValConst
		}
		httpsPortVal, err = strconv.Atoi(httpsPort)
		if err != nil {
			log.Error(err, "Valid https port was not provided. Default port will be used")
			httpsPortVal = httpsPortValConst
		}
	} else {
		verifyHostname = verifyHostNameVal
	}

	//Retrieving configmap related to micro-gateway configuration mustache/template
	confTemplate, confErr := getConfigmap(r, mgwConfMustache, wso2NameSpaceConst)
	if confErr != nil {
		log.Error(err, "error in retrieving the config map ")
	}
	//retrieve micro-gw-conf from the configmap
	confTemp := confTemplate.Data[mgwConfGoTmpl]

	mgwConfValues := &MGWConf{
		KeystorePath:                   keystorePath,
		KeystorePassword:               keystorePassword,
		TruststorePath:                 truststorePath,
		TruststorePassword:             truststorePassword,
		KeymanagerServerurl:            keymanagerServerurl,
		KeymanagerUsername:             keymanagerUsername,
		KeymanagerPassword:             keymanagerPassword,
		JwtConfigs:                     jwtConfigs,
		EnabledGlobalTMEventPublishing: enabledGlobalTMEventPublishing,
		JmsConnectionProvider:          jmsConnectionProvider,
		ThrottleEndpoint:               throttleEndpoint,
		EnableRealtimeMessageRetrieval: enableRealtimeMessageRetrieval,
		EnableRequestValidation:        enableRequestValidation,
		EnableResponseValidation:       enableRequestValidation,
		LogLevel:                       logLevel,
		HttpPort:                       httpPort,
		HttpsPort:                      httpsPort,
		BasicUsername:                  basicUsername,
		BasicPassword:                  basicPassword,
		AnalyticsEnabled:               analyticsEnabled,
		AnalyticsUsername:              analyticsUsername,
		AnalyticsPassword:              analyticsPassword,
		UploadingTimeSpanInMillis:      uploadingTimeSpanInMillis,
		RotatingPeriod:                 rotatingPeriod,
		UploadFiles:                    uploadFiles,
		VerifyHostname:                 verifyHostname,
		Hostname:                       hostname,
		Port:                           port,
	}

	//generate mgw conf from the template
	mgwConftmpl, err := template.New("").Parse(confTemp)
	if err != nil {
		log.Error(err, "error in rendering mgw conf with template")
	}
	builder := &strings.Builder{}
	err = mgwConftmpl.Execute(builder, mgwConfValues)
	if err != nil {
		log.Error(err, "error in generating Dockerfile")
	}
	//creating k8s secret from the rendered mgw-conf file
	output := builder.String()
	errCreateSecret := createMGWSecret(r, output, owner, instance)
	if errCreateSecret != nil {
		log.Error(errCreateSecret, "Error in creating micro-gateway conf secret")
	} else {
		log.Info("Successfully created secret")
	}

	generateK8sArtifactsForMgw := controlConfigData[generatekubernbetesartifactsformgw]
	genArtifacts, errGenArtifacts := strconv.ParseBool(generateK8sArtifactsForMgw)
	if errGenArtifacts != nil {
		log.Error(errGenArtifacts, "error reading value for generate k8s artifacts")
	}
	getResourceReqCPU := controlConfigData[resourceRequestCPU]
	getResourceReqMemory := controlConfigData[resourceRequestMemory]
	getResourceLimitCPU := controlConfigData[resourceLimitCPU]
	getResourceLimitMemory := controlConfigData[resourceLimitMemory]

	analyticsEnabledBool, _ := strconv.ParseBool(analyticsEnabled)
	dep := createMgwDeployment(instance, controlConf, analyticsEnabledBool, r, userNameSpace, owner,
		getResourceReqCPU, getResourceReqMemory, getResourceLimitCPU, getResourceLimitMemory, containerList,
		int32(httpPortVal), int32(httpsPortVal))
	depFound := &appsv1.Deployment{}
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)

	svc := createMgwLBService(r, instance, userNameSpace, owner, int32(httpPortVal), int32(httpsPortVal), operatorMode)
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
	errGettingHpa := createHorizontalPodAutoscaler(dep, r, owner, minReplicas, maxReplicas, targetAvgUtilizationCPU)
	if errGettingHpa != nil {
		log.Error(errGettingHpa, "Error getting HPA")
	}
	kanikoArgs, err = getConfigmap(r, "kaniko-arguments", "wso2-system")
	if err != nil && errors.IsNotFound(err) {
		log.Info("No kaniko-arguments config map is available in wso2-system namespace")
	}

	// copy configs and secrets to user's namespace
	if err = copyConfigVolumes(r, userNameSpace); err != nil {
		log.Error(err, "Error coping registry specific configs to user's namespace", "user's namespace", userNameSpace)
	}

	// append kaniko specific volumes
	tmpVolMounts, tmpVols := getVolumes(instance.Name, swaggerCmNames)
	jobVolumeMount = append(jobVolumeMount, tmpVolMounts...)
	jobVolume = append(jobVolume, tmpVols...)

	if instance.Spec.UpdateTimeStamp != "" {
		//Schedule Kaniko pod
		reqLogger.Info("Updating the API", "API.Name", instance.Name, "API.Namespace", instance.Namespace)
		job := scheduleKanikoJob(instance, controlConf, jobVolumeMount, jobVolume, instance.Spec.UpdateTimeStamp, owner)
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
			// Kaniko completed the push image, so the image is exists but kaniko job is exists
			// delete completed kaniko job if kaniko is exists
			errDeleteJob := deleteCompletedJobs(instance.Namespace)
			if errDeleteJob != nil {
				log.Error(errDeleteJob, "error deleting completed jobs")
			}

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

					reqLogger.Info("Operator mode is set to " + operatorMode)
					if strings.EqualFold(operatorMode, ingressMode) {
						ingErr := createorUpdateMgwIngressResource(r, instance, userNameSpace, int32(httpPortVal),
							int32(httpsPortVal), apiBasePaths, controlIngressConf, owner)
						if ingErr != nil {
							return reconcile.Result{}, ingErr
						}
					}
					if strings.EqualFold(operatorMode, routeMode) {
						rutErr := createMgwRouteResource(r, instance, userNameSpace, int32(httpPortVal),
							int32(httpsPortVal), apiBasePaths, controlOpenshiftConf, owner)
						if rutErr != nil {
							return reconcile.Result{}, rutErr
						}
					}
					//Service created successfully - don't requeue
					return reconcile.Result{}, nil
				} else if svcErr != nil {
					return reconcile.Result{}, svcErr
				}
				// if service already exsits
				reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
					svcFound.Namespace, "SVC.Name", svcFound.Name)
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

		// delete completed kaniko job
		errDeleteJob := deleteCompletedJobs(instance.Namespace)
		if errDeleteJob != nil {
			log.Error(errDeleteJob, "error deleting completed jobs")
		}

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

				reqLogger.Info("Operator mode is set to " + operatorMode)
				if strings.EqualFold(operatorMode, ingressMode) {
					ingErr := createorUpdateMgwIngressResource(r, instance, userNameSpace, int32(httpPortVal),
						int32(httpsPortVal), apiBasePaths, controlIngressConf, owner)
					if ingErr != nil {
						return reconcile.Result{}, ingErr
					}
				}
				if strings.EqualFold(operatorMode, routeMode) {
					rutErr := createMgwRouteResource(r, instance, userNameSpace, int32(httpPortVal),
						int32(httpsPortVal), apiBasePaths, controlOpenshiftConf, owner)
					if rutErr != nil {
						return reconcile.Result{}, rutErr
					}
				}
				//Service created successfully - don't requeue
				return reconcile.Result{}, nil
			} else if svcErr != nil {
				return reconcile.Result{}, svcErr
			}
			// if service already exsits
			reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
				svcFound.Namespace, "SVC.Name", svcFound.Name)
		} else {
			log.Info("skip generating kubernetes artifacts")
		}

		return reconcile.Result{}, nil
	} else {
		//Schedule Kaniko pod
		job := scheduleKanikoJob(instance, controlConf, jobVolumeMount, jobVolume, instance.Spec.UpdateTimeStamp, owner)
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
			// delete completed kaniko job
			// this delete only happened in override scenario
			errDeleteJob := deleteCompletedJobs(instance.Namespace)
			if errDeleteJob != nil {
				log.Error(errDeleteJob, "error deleting completed jobs")
			}

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

					reqLogger.Info("Operator mode is set to " + operatorMode)
					if strings.EqualFold(operatorMode, ingressMode) {
						ingErr := createorUpdateMgwIngressResource(r, instance, userNameSpace, int32(httpPortVal),
							int32(httpsPortVal), apiBasePaths, controlIngressConf, owner)
						if ingErr != nil {
							return reconcile.Result{}, ingErr
						}
					}
					if strings.EqualFold(operatorMode, routeMode) {
						rutErr := createMgwRouteResource(r, instance, userNameSpace, int32(httpPortVal),
							int32(httpsPortVal), apiBasePaths, controlOpenshiftConf, owner)
						if rutErr != nil {
							return reconcile.Result{}, rutErr
						}
					}
					//Service created successfully - don't requeue
					return reconcile.Result{}, nil
				} else if svcErr != nil {
					return reconcile.Result{}, svcErr
				}
				// if service already exsits
				reqLogger.Info("Skip reconcile: Service already exists", "SVC.Namespace",
					svcFound.Namespace, "SVC.Name", svcFound.Name)
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
		if volume.Secret != nil {
			name := volume.Secret.SecretName
			if err := copySecret(r, name, wso2NameSpaceConst, name, namespace); err != nil {
				return err
			}
		}
		if volume.ConfigMap != nil {
			name := volume.ConfigMap.Name
			if err := copyConfigMap(r, name, wso2NameSpaceConst, name, namespace); err != nil {
				return err
			}
		}
	}

	return nil
}

// gets the data from analytics secret
func getSecretData(r *ReconcileAPI, analyticsSecretName string) (map[string][]byte, error) {
	var analyticsData map[string][]byte
	// Check if this secret exists
	analyticsSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: analyticsSecretName, Namespace: wso2NameSpaceConst}, analyticsSecret)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Analytics Secret is not found")
		return analyticsData, err

	} else if err != nil {
		log.Error(err, "error ")
		return analyticsData, err

	}

	analyticsData = analyticsSecret.Data
	log.Info("Analytics Secret exists")
	return analyticsData, nil

}

//Handles microgateway conf create and update
func createMGWSecret(r *ReconcileAPI, confData string, owner []metav1.OwnerReference, cr *wso2v1alpha1.API) error {
	var apimSecret *corev1.Secret

	apimSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.Name + "-" + mgwConfSecretConst,
			Namespace:       cr.Namespace,
			OwnerReferences: owner,
		},
	}

	apimSecret.Data = map[string][]byte{
		mgwConfConst: []byte(confData),
	}

	// Check if mgw-conf secret exists
	checkSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-" + mgwConfSecretConst, Namespace: cr.Namespace}, checkSecret)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating mgw-conf secret ")
		errSecret := r.client.Create(context.TODO(), apimSecret)
		return errSecret
	} else if err != nil {
		log.Error(err, "error in mgw-conf creation")
		return err
	} else {
		log.Info("Updating mgw-conf secret")
		errSecret := r.client.Update(context.TODO(), apimSecret)
		return errSecret
	}
}

func createHorizontalPodAutoscaler(dep *appsv1.Deployment, r *ReconcileAPI, owner []metav1.OwnerReference,
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

//get configmap
func getConfigmap(r *ReconcileAPI, mapName string, ns string) (*corev1.ConfigMap, error) {
	apiConfigMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: mapName, Namespace: ns}, apiConfigMap)

	if mapName == apimConfName {
		if err != nil && errors.IsNotFound(err) {
			logrus.Warnf("missing APIM configurations ", err)
			return nil, err

		} else if err != nil {
			log.Error(err, "error getting APIM configmap", "configmap name", mapName)
			return apiConfigMap, err
		}
	} else {
		if err != nil && errors.IsNotFound(err) {
			log.Error(err, "Specified configmap is not found", "configmap name", mapName)
			return apiConfigMap, err
		} else if err != nil {
			log.Error(err, "error getting configmap", "configmap name", mapName)
			return apiConfigMap, err
		}
	}
	return apiConfigMap, nil
}

// createConfigMap creates a config file with the given data
func createConfigMap(apiConfigMapRef string, key string, value string, ns string, owner []metav1.OwnerReference) *corev1.ConfigMap {

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            apiConfigMapRef,
			Namespace:       ns,
			OwnerReferences: owner,
		},
		Data: map[string]string{
			key: value,
		},
	}
}

// createSecret creates a config file with the given data
func createSecret(secretName string, key string, value string, ns string, owner []metav1.OwnerReference) *corev1.Secret {

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            secretName,
			Namespace:       ns,
			OwnerReferences: owner,
		},
		Data: map[string][]byte{
			key: []byte(value),
		},
	}
}

//Swagger handling
func mgwSwaggerLoader(swaggerDataMap map[string]string) (*openapi3.Swagger, string, error) {
	var swaggerData string
	var swaggerDataFile string
	for key, value := range swaggerDataMap {
		swaggerData = value
		swaggerDataFile = key
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(swaggerData))

	swaggerV3Version := swagger.OpenAPI
	log.Info("Swagger version", swaggerV3Version)

	if swaggerV3Version != "" {
		return swagger, swaggerDataFile, err
	} else {
		log.Info("OpenAPI v3 not found. Hence converting Swagger 2 to Swagger 3")
		var swagger2 openapi2.Swagger
		err2 := yaml.Unmarshal([]byte(swaggerData), &swagger2)
		swaggerV3, err2 := openapi2conv.ToV3Swagger(&swagger2)
		return swaggerV3, swaggerDataFile, err2
	}
}

//Get endpoint from swagger and replace it with targetendpoint kind service endpoint
func mgwSwaggerHandler(r *ReconcileAPI, swagger *openapi3.Swagger, mode string, userNameSpace string) (map[string]string, *openapi3.Swagger, string) {

	var mgwSwagger *openapi3.Swagger

	mgwSwagger = swagger
	endpointNames := make(map[string]string)
	var checkt []string
	var apiBasePath string
	//api level endpoint
	endpointData, checkEndpoint := swagger.Extensions[endpointExtension]
	if checkEndpoint {
		prodEp := XMGWProductionEndpoints{}
		var endPoint string
		endpointJson, checkJsonRaw := endpointData.(json.RawMessage)
		if checkJsonRaw {
			err := json.Unmarshal(endpointJson, &endPoint)
			if err == nil {
				log.Info("Parsing endpoints and not available root service endpoint")
				//check if service & targetendpoint cr object are available
				extractData := strings.Split(endPoint, ".")
				if len(extractData) == 2 {
					userNameSpace = extractData[1]
					endPoint = extractData[0]
				}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: endPoint}, targetEndpointCr)

				if erCr != nil && errors.IsNotFound(erCr) {
					log.Error(err, "targetEndpoint CRD object is not found")
				} else if erCr != nil {
					log.Error(err, "Error in getting targetendpoint CRD object")
				}
				if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), serverless) {
					currentService := &v1.Service{}
					err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
						Name: endPoint}, currentService)
				} else {
					currentService := &corev1.Service{}
					err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
						Name: endPoint}, currentService)
				}
				if err != nil && errors.IsNotFound(err) && mode != sidecar {
					log.Error(err, "service not found")
				} else if err != nil && mode != sidecar {
					log.Error(err, "Error in getting service")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					if mode == sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						endpointNames[targetEndpointCr.Name] = endPointSidecar
						checkt = append(checkt, endPointSidecar)
					}
					if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), serverless) {

						endPoint = protocol + "://" + endPoint + "." + userNameSpace + ".svc.cluster.local"
						checkt = append(checkt, endPoint)

					} else {
						endPoint = protocol + "://" + endPoint + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						checkt = append(checkt, endPoint)
					}
					prodEp.Urls = checkt
					mgwSwagger.Extensions[endpointExtension] = prodEp
				}
			} else {
				err := json.Unmarshal(endpointJson, &prodEp)
				if err == nil {
					lengthOfUrls := len(prodEp.Urls)
					endpointList := make([]string, lengthOfUrls)
					isServiceDef := false
					for index, urlVal := range prodEp.Urls {
						endpointUrl, err := url.Parse(urlVal)
						if err != nil {
							currentService := &corev1.Service{}
							targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
							err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
								Name: urlVal}, currentService)
							erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: urlVal}, targetEndpointCr)
							if err == nil && erCr == nil {
								protocol := targetEndpointCr.Spec.Protocol
								urlVal = protocol + "://" + urlVal + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
								if mode == sidecar {
									urlValSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
									endpointNames[urlVal] = urlValSidecar
									endpointList[index] = urlValSidecar
								} else {
									endpointList[index] = urlVal
								}
								isServiceDef = true
							}
						} else {
							endpointNames[endpointUrl.Hostname()] = endpointUrl.Hostname()
						}
					}

					if isServiceDef {
						prodEp.Urls = endpointList
						mgwSwagger.Extensions[endpointExtension] = prodEp
					}
				} else {
					log.Info("error unmarshal endpoint")
				}
			}
		}
	}

	//extract base path
	basePathData, checkBasePath := swagger.Extensions[apiBasePathExtention]
	if checkBasePath {
		basePathJson, checkJsonRaw := basePathData.(json.RawMessage)
		if checkJsonRaw {
			err := json.Unmarshal(basePathJson, &apiBasePath)
			if err != nil {
				log.Info("Error unmarshal base path")
			}
		}
	}

	//resource level endpoint
	for pathName, path := range swagger.Paths {

		if path.Get != nil {
			getEp, gcep := path.Get.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, getEp, endpointNames, gcep, userNameSpace, mode)
			assignGetEps(mgwSwagger, eps)
		}
		if path.Post != nil {
			postEp, pocep := path.Post.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, postEp, endpointNames, pocep, userNameSpace, mode)
			assignPostEps(mgwSwagger, eps)
		}
		if path.Put != nil {
			putEp, pucep := path.Put.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, putEp, endpointNames, pucep, userNameSpace, mode)
			assignPutEps(mgwSwagger, eps)
		}
		if path.Delete != nil {
			deleteEp, dcep := path.Delete.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, deleteEp, endpointNames, dcep, userNameSpace, mode)
			assignDeleteEps(mgwSwagger, eps)
		}
		if path.Patch != nil {
			pEp, pAvl := path.Patch.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, pEp, endpointNames, pAvl, userNameSpace, mode)
			assignPatchEps(mgwSwagger, eps)
		}
		if path.Head != nil {
			pEp, pAvl := path.Head.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, pEp, endpointNames, pAvl, userNameSpace, mode)
			assignHeadEps(mgwSwagger, eps)
		}
		if path.Options != nil {
			pEp, pAvl := path.Options.Extensions[endpointExtension]
			eps := resolveEps(r, pathName, pEp, endpointNames, pAvl, userNameSpace, mode)
			assignOptionsEps(mgwSwagger, eps)
		}
	}
	return endpointNames, mgwSwagger, apiBasePath
}

func assignGetEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Get.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignPutEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Put.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignPostEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Post.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignDeleteEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Delete.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignPatchEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Patch.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignHeadEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Head.Extensions[endpointExtension] = value
			}
		}
	}
}

func assignOptionsEps(swagger *openapi3.Swagger, resLevelEp map[string]XMGWProductionEndpoints) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Options.Extensions[endpointExtension] = value
			}
		}
	}
}

func resolveEps(r *ReconcileAPI, pathName string, resourceGetEp interface{}, endpointNames map[string]string, checkResourceEP bool,
	userNameSpace string, mode string) map[string]XMGWProductionEndpoints {
	var checkr []string
	var resLevelEp = make(map[string]XMGWProductionEndpoints)

	//resourceGetEp, checkResourceEP := path.Get.Extensions[endpointExtension]
	if checkResourceEP {
		prodEp := XMGWProductionEndpoints{}
		var endPoint string
		ResourceEndpointJson, checkJsonResource := resourceGetEp.(json.RawMessage)
		if checkJsonResource {
			err := json.Unmarshal(ResourceEndpointJson, &endPoint)
			if err == nil {
				//check if service & targetendpoint cr object are available
				currentService := &corev1.Service{}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
					Name: endPoint}, currentService)
				erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: endPoint}, targetEndpointCr)
				if err != nil && errors.IsNotFound(err) && mode != sidecar {
					log.Error(err, "Service is not found")
				} else if erCr != nil && errors.IsNotFound(erCr) {
					log.Error(err, "targetendpoint CRD object is not found")
				} else if err != nil && mode != sidecar {
					log.Error(err, "Error in getting service")
				} else if erCr != nil {
					log.Error(err, "Error in getting targetendpoint CRD object")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					if mode == sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						endpointNames[endPoint] = endPointSidecar
						checkr = append(checkr, endPointSidecar)
					} else {
						endPoint = protocol + "://" + endPoint + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						checkr = append(checkr, endPoint)
					}
					prodEp.Urls = checkr
					resLevelEp[pathName] = prodEp
				}
			} else {
				err := json.Unmarshal(ResourceEndpointJson, &prodEp)
				if err == nil {
					lengthOfUrls := len(prodEp.Urls)
					endpointList := make([]string, lengthOfUrls)
					isServiceDef := false
					for index, urlVal := range prodEp.Urls {
						endpointUrl, err := url.Parse(urlVal)
						if err != nil {
							currentService := &corev1.Service{}
							targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
							err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
								Name: urlVal}, currentService)
							erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: urlVal}, targetEndpointCr)
							if err == nil && erCr == nil || mode == sidecar {
								endpointNames[urlVal] = urlVal
								protocol := targetEndpointCr.Spec.Protocol
								if mode == sidecar {
									urlValSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
									endpointNames[urlVal] = urlValSidecar
									endpointList[index] = urlValSidecar
								} else {
									urlVal = protocol + "://" + urlVal + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
									endpointList[index] = urlVal
								}
								isServiceDef = true
							}
						} else {
							endpointNames[endpointUrl.Hostname()] = endpointUrl.Hostname()
						}
					}

					if isServiceDef {
						prodEp.Urls = endpointList
						resLevelEp[pathName] = prodEp
					}
				}
			}
		}
	}
	return resLevelEp
}

func getCredentials(r *ReconcileAPI, name string, securityType string, userNameSpace string) error {

	hasher := sha1.New()
	var usrname string
	var password []byte

	//get the secret included credentials
	credentialSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: userNameSpace}, credentialSecret)

	if err != nil && errors.IsNotFound(err) {
		log.Info("secret not found")
		return err
	}

	//get the username and the password
	for k, v := range credentialSecret.Data {
		if strings.EqualFold(k, "username") {
			usrname = string(v)
		}
		if strings.EqualFold(k, "password") {
			password = v
		}

	}
	if securityType == "Basic" {

		basicUsername = usrname
		_, err := hasher.Write([]byte(password))
		if err != nil {
			log.Info("error in encoding password")
			return err
		}
		//convert encoded password to a uppercase hex string
		basicPassword = hex.EncodeToString(hasher.Sum(nil))
	}
	if securityType == "Oauth" {
		keymanagerUsername = usrname
		keymanagerPassword = string(password)
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
			APIVersion: apiVersion,
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

//Handles dockerfile configmap creation
func dockerfileHandler(r *ReconcileAPI, certList map[string]string, existcert bool, conf map[string]string,
	owner []metav1.OwnerReference, cr *wso2v1alpha1.API, existInterceptors bool, existJavaInterceptors bool) (*corev1.ConfigMap, error) {
	var dockerTemplate string
	truststorePass := getTruststorePassword(r)
	dockerTemplateConfigmap, err := getConfigmap(r, dockerFileTemplate, wso2NameSpaceConst)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "docker template configmap not found")
		return nil, err
	} else if err != nil {
		log.Error(err, "error in retrieving docker template")
		return nil, err
	}
	for _, val := range dockerTemplateConfigmap.Data {
		dockerTemplate = string(val)
	}
	certs := &DockerfileArtifacts{
		CertFound:             existcert,
		Password:              truststorePass,
		Certs:                 certList,
		BaseImage:             conf[mgwToolkitImgConst],
		RuntimeImage:          conf[mgwRuntimeImgConst],
		InterceptorsFound:     existInterceptors,
		JavaInterceptorsFound: existJavaInterceptors,
	}
	//generate dockerfile from the template
	tmpl, err := template.New("").Parse(dockerTemplate)
	if err != nil {
		log.Error(err, "error in rendering Dockerfile with template")
		return nil, err
	}
	builder := &strings.Builder{}
	err = tmpl.Execute(builder, certs)
	if err != nil {
		log.Error(err, "error in generating Dockerfile")
		return nil, err
	}

	dockerfileConfmap, err := getConfigmap(r, cr.Name+"-"+dockerFile, cr.Namespace)
	data := builder.String()
	if err != nil && errors.IsNotFound(err) {
		dockerConf := createConfigMap(cr.Name+"-"+dockerFile, "Dockerfile", data, cr.Namespace, owner)

		errorMap := r.client.Create(context.TODO(), dockerConf)
		if errorMap != nil {
			return dockerfileConfmap, errorMap
		}
		return dockerConf, nil
	} else if err != nil {
		return dockerfileConfmap, err
	}
	//update existing dockerfile
	dockerfileConfmap.Data["Dockerfile"] = builder.String()
	errorupdate := r.client.Update(context.TODO(), dockerfileConfmap)
	if errorupdate != nil {
		log.Error(errorupdate, "error in updating config map")
	}

	return dockerfileConfmap, err
}

func policyHandler(r *ReconcileAPI, operatorOwner []metav1.OwnerReference, userNameSpace string) error {
	//Check if policy configmap is available
	foundmapc := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: policyConfigmap, Namespace: userNameSpace}, foundmapc)

	if err != nil && errors.IsNotFound(err) {
		//create new map with default policies in user namespace if a map is not found
		log.Info("Creating a config map with default policies", "Namespace", userNameSpace, "Name", policyConfigmap)

		defaultval := ratelimiting.CreateDefault()
		confmap := createConfigMap(policyConfigmap, policyFileConst, defaultval, userNameSpace, operatorOwner)

		err = r.client.Create(context.TODO(), confmap)
		if err != nil {
			log.Error(err, "error ")
			return err
		}
	} else if err != nil {
		log.Error(err, "error ")
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
	dockerConfigSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, dockerConfigSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Docker credentials secret is not found", "secret-name", secretName, "namespace", namespace)
	} else if err != nil {
		log.Error(err, "Error while getting docker credentials secret", "secret-name", secretName, "namespace", namespace)
	} else {
		authsJsonString := dockerConfigSecret.Data[utils.DockerConfigKeyConst]
		auths := Auth{}
		err := json.Unmarshal([]byte(authsJsonString), &auths)
		if err != nil {
			log.Info("Error unmarshal data of docker credential auth")
		}

		for regUrl, credential := range auths.Auths {
			registryUrl = removeVersionTag(regUrl)
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
	jobVolume []corev1.Volume, timeStamp string, owner []metav1.OwnerReference) *batchv1.Job {
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
			OwnerReferences: owner,
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

// Creating a LB balancer service to expose mgw
// Supports for multiple apiBasePaths when there are multiple swaggers for one API CRD
func createorUpdateMgwIngressResource(r *ReconcileAPI, cr *wso2v1alpha1.API, nameSpace string, httpPortVal int32,
	httpsPortVal int32, apiBasePaths []string, controllerConfig *corev1.ConfigMap, owner []metav1.OwnerReference) error {
	controlConfigData := controllerConfig.Data
	transportMode := controlConfigData[ingressTransportMode]
	ingressHostName := controlConfigData[ingressHostName]
	tlsSecretName := controlConfigData[tlsSecretName]
	ingressName := controlConfigData[ingressResourceName]

	var hostArray []string
	hostArray = append(hostArray, ingressHostName)
	log.Info(fmt.Sprintf("Creating ingress resource with name: %v", ingressName))
	log.Info(fmt.Sprintf("Creating ingress resource with API Base Path: %v", apiBasePaths))
	log.WithValues("Ingress metadata. Transport mode", transportMode, "Ingress name", ingressName,
		"Ingress hostname "+ingressHostName)
	annotationMap, err := getConfigmap(r, ingressConfigs, wso2NameSpaceConst)
	var port int32

	if httpConst == transportMode {
		port = httpPortVal
	} else {
		port = httpsPortVal
	}

	apiServiceName := cr.Name
	ingress := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingressName, Namespace: nameSpace}, ingress)
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

	// add multiple api base paths
	var httpIngressPaths []v1beta1.HTTPIngressPath
	for _, apiBasePath := range apiBasePaths {
		httpIngressPaths = append(httpIngressPaths, v1beta1.HTTPIngressPath{
			Path: apiBasePath,
			Backend: v1beta1.IngressBackend{
				ServiceName: apiServiceName,
				ServicePort: intstr.IntOrString{IntVal: port},
			},
		})
	}

	log.Info(fmt.Sprintf("Creating ingress resource with name: %v", ingressName))
	if err != nil && errors.IsNotFound(err) {
		log.Info("Ingress resource not found with name" + ingressName + ".Hence creating a new ingress resource")
		ingress := &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       nameSpace, // goes into backend full name
				Name:            ingressName + "-" + cr.Name,
				Annotations:     ingressAnnotationMap,
				OwnerReferences: owner,
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
		err = r.client.Create(context.TODO(), ingress)
		return err
	} else {
		log.Info("Ingress resource found with name" + ingressName + ".Hence updating the existing ingress resource")
		rules := ingress.Spec.Rules
		var rulesArray []v1beta1.IngressRule
		var update bool = false
		for _, element := range rules {
			var pathArray []v1beta1.HTTPIngressPath
			for _, path := range element.IngressRuleValue.HTTP.Paths {
				if isStringArrayContains(apiBasePaths, path.Path) {
					path.Backend.ServiceName = apiServiceName
					path.Backend.ServicePort = intstr.IntOrString{IntVal: port}
					update = true
				}
				pathArray = append(pathArray, path)
				element.IngressRuleValue.HTTP.Paths = pathArray
			}
			rulesArray = append(rulesArray, element)
		}

		if update {
			log.Info(fmt.Sprintf("Ingress API Base path found with name %v.Hence updating the rule", apiBasePaths))
			ingress.Spec.Rules = rulesArray
			err = r.client.Update(context.TODO(), ingress)
			return err
		}

		rulesArray = make([]v1beta1.IngressRule, 0)
		for _, element := range rules {
			paths := element.IngressRuleValue.HTTP.Paths
			paths = append(paths, httpIngressPaths...)
			element.IngressRuleValue.HTTP.Paths = paths
			rulesArray = append(rulesArray, element)
			ingress.Spec.Rules = rulesArray
			err = r.client.Update(context.TODO(), ingress)
			return err
		}
	}
	return err
}

func createMgwRouteResource(r *ReconcileAPI, cr *wso2v1alpha1.API, nameSpace string, httpPortVal int32,
	httpsPortVal int32, apiBasePaths []string, controllerConfig *corev1.ConfigMap, owner []metav1.OwnerReference) error {

	controlConfigData := controllerConfig.Data
	routeName := controlConfigData[routeName]
	routeHost := controlConfigData[routeHost]
	routeTransportMode := controlConfigData[routeTransportMode]

	route := &routv1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: routeName, Namespace: nameSpace}, route)
	var port int32

	if httpConst == routeTransportMode {
		port = httpPortVal
	} else {
		port = httpsPortVal
	}

	log.Info(fmt.Sprintf("Creating Route resource with name: %v", routeName))
	if err != nil && errors.IsNotFound(err) {
		log.Info("Route resource not found with name" + routeName + ".Hence creating a new route resource")
		serviceName := cr.Name
		route := &routv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:            cr.Name,
				Namespace:       nameSpace,
				OwnerReferences: owner,
			},
			Spec: routv1.RouteSpec{
				Host: routeHost,
				Path: apiBasePaths[0],
				Port: &routv1.RoutePort{
					TargetPort: intstr.IntOrString{IntVal: port},
				},
				To: routv1.RouteTargetReference{
					Kind: "Service",
					Name: serviceName,
				},
			},
		}
		err = r.client.Create(context.TODO(), route)
	}
	return err
}

//default volume mounts for the kaniko job
func getVolumes(apiName string, swaggerCmNames []string) ([]corev1.VolumeMount, []corev1.Volume) {
	regConfig := registry.GetConfig()

	jobVolumeMount := []corev1.VolumeMount{
		{
			Name:      mgwDockerFile,
			MountPath: dockerFileLocation,
		},
		{
			Name:      policyyamlFile,
			MountPath: policyyamlLocation,
			ReadOnly:  true,
		},
		{
			Name:      mgwConfFile,
			MountPath: mgwConfLocation,
			ReadOnly:  true,
		},
	}
	// append secrets from regConfig
	jobVolumeMount = append(jobVolumeMount, regConfig.VolumeMounts...)

	jobVolume := []corev1.Volume{
		{
			Name: mgwDockerFile,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiName + "-" + dockerFile,
					},
				},
			},
		},
		{
			Name: policyyamlFile,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: policyConfigmap,
					},
				},
			},
		},
		{
			Name: mgwConfFile,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: apiName + "-" + mgwConfSecretConst,
				},
			},
		},
	}
	// append secrets from regConfig
	jobVolume = append(jobVolume, regConfig.Volumes...)

	// append swagger file config maps
	for i, swaggerCmName := range swaggerCmNames {
		jobVolume = append(jobVolume, corev1.Volume{
			Name: swaggerCmName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: swaggerCmName + "-mgw",
					},
				},
			},
		})

		jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
			Name:      swaggerCmName,
			ReadOnly:  true,
			MountPath: fmt.Sprintf(swaggerLocation, i+1),
		})
	}

	return jobVolumeMount, jobVolume
}

// Handles the mounting of analytics certificate
func analyticsVolumeHandler(analyticsCertSecretName string, r *ReconcileAPI, jobVolumeMount []corev1.VolumeMount,
	jobVolume []corev1.Volume, userNameSpace string, operatorOwner []metav1.OwnerReference) ([]corev1.VolumeMount, []corev1.Volume, string, error) {
	var fileName string
	var value string
	analyticsCertSecret := &corev1.Secret{}
	//checks if the certificate exists in the user namepspace
	errCertNs := r.client.Get(context.TODO(), types.NamespacedName{Name: analyticsCertSecretName, Namespace: userNameSpace}, analyticsCertSecret)

	if errCertNs != nil {
		log.Info("Error in getting certificate secret specified in analytics from the user namespace. Finding it in " + wso2NameSpaceConst)
		errCert := r.client.Get(context.TODO(), types.NamespacedName{Name: analyticsCertSecretName, Namespace: wso2NameSpaceConst}, analyticsCertSecret)
		if errCert != nil {
			log.Error(errCert, "Error in getting certificate secret specified in analytics from "+wso2NameSpaceConst)
			return jobVolumeMount, jobVolume, fileName, errCert
		}
		for pem, val := range analyticsCertSecret.Data {
			fileName = pem
			value = string(val)
		}
		newSecret := createSecret(analyticsCertSecretName, fileName, value, userNameSpace, operatorOwner)
		err := r.client.Create(context.TODO(), newSecret)
		if err != nil {
			log.Error(err, "Error in copying analytics cert to user namespace")
			return jobVolumeMount, jobVolume, fileName, err
		}
		log.Info("Successfully copied analytics cert to user namespace")
	}
	log.Info("Mounting analytics cert to volume.")
	jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
		Name:      analyticsCertFile,
		MountPath: analyticsCertLocation,
		ReadOnly:  true,
	})
	jobVolume = append(jobVolume, corev1.Volume{
		Name: analyticsCertFile,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: analyticsCertSecretName,
			},
		},
	})
	for pem := range analyticsCertSecret.Data {
		fileName = pem
	}
	return jobVolumeMount, jobVolume, fileName, nil
}

func certMoutHandler(r *ReconcileAPI, cert *corev1.Secret, jobVolumeMount []corev1.VolumeMount, jobVolume []corev1.Volume) ([]corev1.VolumeMount, []corev1.Volume) {
	name := certConfig + "-" + cert.Name
	// check volume already exists
	for _, volume := range jobVolume {
		if volume.Name == name {
			return jobVolumeMount, jobVolume
		}
	}

	jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
		Name:      name,
		MountPath: certPath + cert.Name,
		ReadOnly:  true,
	})

	jobVolume = append(jobVolume, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: cert.Name,
			},
		},
	})
	return jobVolumeMount, jobVolume
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

func getTruststorePassword(r *ReconcileAPI) string {

	var password string
	//get secret if available
	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: truststoreSecretName, Namespace: wso2NameSpaceConst},
		secret)
	if err != nil && errors.IsNotFound(err) {
		encodedpassword := encodedTrustsorePassword
		//decode and get the password to append to the dockerfile
		decodedpass, err := b64.StdEncoding.DecodeString(encodedpassword)
		if err != nil {
			log.Error(err, "error decoding truststore password")
		}
		password = string(decodedpass)
		log.Info("creating new secret for truststore password")
		var truststoresecret *corev1.Secret
		//create a new secret with password
		truststoresecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      truststoreSecretName,
				Namespace: wso2NameSpaceConst,
			},
		}
		truststoresecret.Data = map[string][]byte{
			truststoreSecretData: []byte(encodedpassword),
		}
		errsecret := r.client.Create(context.TODO(), truststoresecret)
		log.Error(errsecret, "error in creating trustore password")
		return password
	}
	//get password from the secret
	foundpassword := string(secret.Data[truststoreSecretData])
	getpass, err := b64.StdEncoding.DecodeString(foundpassword)
	if err != nil {
		log.Error(err, "error decoding truststore password")
	}
	password = string(getpass)
	return password
}

//gets the details of the api crd object for owner reference
func getOwnerDetails(cr *wso2v1alpha1.API) []metav1.OwnerReference {
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

//gets the details of the operator for owner reference
func getOperatorOwner(r *ReconcileAPI) ([]metav1.OwnerReference, error) {
	depFound := &appsv1.Deployment{}
	setOwner := true
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: "api-operator", Namespace: wso2NameSpaceConst}, depFound)
	if deperr != nil {
		noOwner := []metav1.OwnerReference{}
		return noOwner, deperr
	}
	return []metav1.OwnerReference{
		{
			APIVersion:         depFound.APIVersion,
			Kind:               depFound.Kind,
			Name:               depFound.Name,
			UID:                depFound.UID,
			Controller:         &setOwner,
			BlockOwnerDeletion: &setOwner,
		},
	}, nil
}

func getSecurityDefinedInSwagger(swagger *openapi3.Swagger) (map[string][]string, bool, int, error) {

	//get all the securities defined in swagger
	var securityMap = make(map[string][]string)
	var securityDef path
	//get API level security
	apiLevelSecurity, isDefined := swagger.Extensions[securityExtension]
	var APILevelSecurity []map[string][]string
	if isDefined {
		log.Info("API level security is defined")
		rawmsg := apiLevelSecurity.(json.RawMessage)
		errsec := json.Unmarshal(rawmsg, &APILevelSecurity)
		if errsec != nil {
			log.Error(errsec, "error unmarshaling API level security ")
			return securityMap, isDefined, len(securityDef.Security), errsec
		}
		for _, value := range APILevelSecurity {
			for secName, val := range value {
				securityMap[secName] = val
			}
		}
	} else {
		log.Info("API Level security is not defined")
	}
	//get resource level security
	resLevelSecurity, resSecIsDefined := swagger.Extensions[pathsExtension]
	var resSecurityMap map[string]map[string]path

	if resSecIsDefined {
		rawmsg := resLevelSecurity.(json.RawMessage)
		errrSec := json.Unmarshal(rawmsg, &resSecurityMap)
		if errrSec != nil {
			log.Error(errrSec, "error unmarshall into resource level security")
			return securityMap, isDefined, len(securityDef.Security), errrSec
		}
		for _, path := range resSecurityMap {
			for _, sec := range path {
				securityDef = sec

			}
		}
	}
	if len(securityDef.Security) > 0 {
		log.Info("Resource level security is defined")
		for _, obj := range resSecurityMap {
			for _, obj := range obj {
				for _, value := range obj.Security {
					for secName, val := range value {
						securityMap[secName] = val
					}
				}
			}
		}
	} else {
		log.Info("Resource level security is not defiend")
	}
	return securityMap, isDefined, len(securityDef.Security), nil
}

func handleSecurity(r *ReconcileAPI, securityMap map[string][]string, userNameSpace string, instance *wso2v1alpha1.API, secSchemeDefined bool, certList map[string]string, jobVolumeMount []corev1.VolumeMount, jobVolume []corev1.Volume) (map[string]securitySchemeStruct, bool, map[string]string, []corev1.VolumeMount, []corev1.Volume, []SecurityTypeJWT, error) {

	var alias string
	//keep to track the existence of certificates
	var existSecCert bool

	var securityDefinition = make(map[string]securitySchemeStruct)
	//to add multiple certs with alias

	var certificateName string

	jwtConfArray := []SecurityTypeJWT{}
	securityInstance := &wso2v1alpha1.Security{}
	var certificateSecret = &corev1.Secret{}
	for secName, scopeList := range securityMap {
		//retrieve security instances
		errGetSec := r.client.Get(context.TODO(), types.NamespacedName{Name: secName, Namespace: userNameSpace}, securityInstance)
		if errGetSec != nil && errors.IsNotFound(errGetSec) {
			log.Info("defined security instance " + secName + " is not found")
			return securityDefinition, existSecCert, certList, jobVolumeMount, jobVolume, jwtConfArray, errGetSec
		}
		if strings.EqualFold(securityInstance.Spec.Type, securityOauth) {
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				errc := r.client.Get(context.TODO(), types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
				if errc != nil && errors.IsNotFound(errc) {
					log.Info("defined certificate is not found")
					return securityDefinition, existSecCert, certList, jobVolumeMount, jobVolume, jwtConfArray, errc
				} else {
					log.Info("defined certificate successfully retrieved")
				}
				//mount certs
				volumemountTemp, volumeTemp := certMoutHandler(r, certificateSecret, jobVolumeMount, jobVolume)
				jobVolumeMount = volumemountTemp
				jobVolume = volumeTemp
				alias = certificateSecret.Name + certAlias
				existSecCert = true
				for k := range certificateSecret.Data {
					certificateName = k
				}
				//add cert path and alias as key value pairs
				certList[alias] = certPath + certificateSecret.Name + "/" + certificateName
				//get the keymanager server URL from the security kind
				keymanagerServerurl = securityConf.Endpoint
				//fetch credentials from the secret created
				errGetCredentials := getCredentials(r, securityConf.Credentials, securityOauth, userNameSpace)
				if errGetCredentials != nil {
					log.Error(errGetCredentials, "Error occurred when retrieving credentials for Oauth")
				} else {
					log.Info("Credentials successfully retrieved for security " + secName)
				}
				if !secSchemeDefined {
					//add scopes
					scopes := map[string]string{}
					for _, scopeValue := range scopeList {
						scopes[scopeValue] = "grant " + scopeValue + " access"
					}
					//creating security scheme
					scheme := securitySchemeStruct{
						SecurityType: oauthSecurityType,
						Flows: &authorizationCode{
							scopeSet{
								authorizationUrl,
								tokenUrl,
								scopes,
							},
						},
					}
					securityDefinition[secName] = scheme
				}
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, securityJWT) {
			log.Info("retrieving data for security type JWT")
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				jwtConf := SecurityTypeJWT{}
				errc := r.client.Get(context.TODO(), types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
				if errc != nil && errors.IsNotFound(errc) {
					log.Info("defined certificate is not found")
					return securityDefinition, existSecCert, certList, jobVolumeMount, jobVolume, jwtConfArray, errc
				} else {
					log.Info("defined certificate successfully retrieved")
				}
				//mount certs
				volumemountTemp, volumeTemp := certMoutHandler(r, certificateSecret, jobVolumeMount, jobVolume)
				jobVolumeMount = volumemountTemp
				jobVolume = volumeTemp
				alias = certificateSecret.Name + certAlias
				existSecCert = true
				for k := range certificateSecret.Data {
					certificateName = k
				}
				//add cert path and alias as key value pairs
				certList[alias] = certPath + certificateSecret.Name + "/" + certificateName
				log.Info("certificate alias", alias)
				jwtConf.CertificateAlias = alias
				jwtConf.ValidateSubscription = securityConf.ValidateSubscription

				if securityConf.Issuer != "" {
					jwtConf.Issuer = securityConf.Issuer
				}
				if securityConf.Audience != "" {
					jwtConf.Audience = securityConf.Audience
				}

				log.Info("certificate issuer", issuer)
				jwtConfArray = append(jwtConfArray, jwtConf)
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, basicSecurityAndScheme) {
			// "existCert = false" for this scenario and do not change the global "existCert" value
			// i.e. if global "existCert" is true, even though the scenario for this swagger is false keep that value as true

			//fetch credentials from the secret created
			errGetCredentials := getCredentials(r, securityInstance.Spec.SecurityConfig[0].Credentials, "Basic", userNameSpace)
			if errGetCredentials != nil {
				log.Error(errGetCredentials, "Error occurred when retrieving credentials for Basic")
			} else {
				log.Info("Credentials successfully retrieved for security " + secName)
			}
			//creating security scheme
			if !secSchemeDefined {
				scheme := securitySchemeStruct{
					SecurityType: basicSecurityType,
					Scheme:       basicSecurityAndScheme,
				}
				securityDefinition[secName] = scheme
			}
		}
	}
	return securityDefinition, existSecCert, certList, jobVolumeMount, jobVolume, jwtConfArray, nil
}

func copyDefaultSecurity(securityDefault *wso2v1alpha1.Security, userNameSpace string, owner []metav1.OwnerReference) *wso2v1alpha1.Security {

	securityConf := wso2v1alpha1.SecurityConfig{
		Certificate: securityDefault.Spec.SecurityConfig[0].Certificate,
		Audience:    securityDefault.Spec.SecurityConfig[0].Audience,
		Issuer:      securityDefault.Spec.SecurityConfig[0].Issuer,
	}

	securityConfArray := []wso2v1alpha1.SecurityConfig{}

	securityConfArray = append(securityConfArray, securityConf)
	return &wso2v1alpha1.Security{
		ObjectMeta: metav1.ObjectMeta{
			Name:            defaultSecurity,
			Namespace:       userNameSpace,
			OwnerReferences: owner,
		},
		Spec: wso2v1alpha1.SecuritySpec{
			Type:           securityDefault.Spec.Type,
			SecurityConfig: securityConfArray,
		},
	}
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

//update configmaps with OwnerReference
func updateConfMapWithOwner(r *ReconcileAPI, owner []metav1.OwnerReference, configMap *corev1.ConfigMap) error {
	configMap.OwnerReferences = owner
	errorUpdateinterceptConf := r.client.Update(context.TODO(), configMap)
	if errorUpdateinterceptConf != nil {
		return errorUpdateinterceptConf
	}
	return nil
}

//Hanldling interceptors to modify request and response flows
func interceptorHandler(r *ReconcileAPI, instance *wso2v1alpha1.API, owner []metav1.OwnerReference,
	jobVolumeMount []corev1.VolumeMount, jobVolume []corev1.Volume, userNameSpace string) (bool, bool, []corev1.VolumeMount, []corev1.Volume, error, error) {

	//values to return
	var exsistBalInterceptors bool
	var errBalInterceptor error

	//handle bal interceptors
	balConfigs := instance.Spec.Definition.Interceptors.Ballerina
	for i, balConfig := range balConfigs {
		interceptorConfigmap, err := getConfigmap(r, balConfig, userNameSpace)
		if err != nil {
			if errors.IsNotFound(err) {
				// Interceptors are not defined
				log.Info("ballerina interceptors are not defined")
				exsistBalInterceptors = false
				errBalInterceptor = nil
			} else {
				// Error getting interceptors configmap.
				log.Error(err, "error retrieving ballerina interceptors configmap "+instance.Name+"-interceptors")
				exsistBalInterceptors = false
				errBalInterceptor = err
			}
		} else {
			//mount interceptors configmap to the volume
			log.Info("Mounting interceptors configmap to volume.")
			name := fmt.Sprintf("%s-%s", balConfig, interceptorsVolume)
			jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
				Name:      name,
				MountPath: fmt.Sprintf(interceptorsVolumeLocation, i),
				ReadOnly:  true,
			})
			jobVolume = append(jobVolume, corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: balConfig,
						},
					},
				},
			})
			//update configmap with owner reference
			log.Info("updating interceptors configmap with owner reference")
			errorUpdateinterceptConf := updateConfMapWithOwner(r, owner, interceptorConfigmap)
			if errorUpdateinterceptConf != nil {
				log.Error(errorUpdateinterceptConf, "error in updating interceptors config map with owner reference")
			}
			exsistBalInterceptors = true
			errBalInterceptor = nil
		}
	}

	//handle java interceptors
	var confNames = instance.Spec.Definition.Interceptors.Java
	if len(confNames) > 0 {
		log.Info("java interceptor configmaps specified in API spec")
		for i, configmapName := range confNames {
			javaConfigmap, err := getConfigmap(r, configmapName, userNameSpace)
			if err != nil {
				if errors.IsNotFound(err) {
					// Interceptor is not defined
					log.Info("interceptor" + configmapName + " is not defined")
					return exsistBalInterceptors, false, jobVolumeMount, jobVolume, errBalInterceptor, nil
				} else {
					// Error getting interceptors configmap.
					log.Error(err, "error retrieving configmap "+configmapName)
					return exsistBalInterceptors, false, jobVolumeMount, jobVolume, errBalInterceptor, err
				}
			} else {
				volName := fmt.Sprintf("%s-%s", configmapName, javaInterceptorsVolume)
				jobVolume = append(jobVolume, corev1.Volume{
					Name: volName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configmapName,
							},
						},
					},
				})
				jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
					Name:      volName,
					MountPath: fmt.Sprintf(javaInterceptorsVolumeLocation, i),
					ReadOnly:  true,
				})
			}
			//update configmap with owner reference
			log.Info("updating java interceptor configmap" + configmapName + " with owner reference")
			errorUpdateinterceptConf := updateConfMapWithOwner(r, owner, javaConfigmap)
			if errorUpdateinterceptConf != nil {
				log.Error(errorUpdateinterceptConf, "error in updating java-interceptor configmap"+configmapName+" with owner reference")
			}
		}
		return exsistBalInterceptors, true, jobVolumeMount, jobVolume, errBalInterceptor, nil
	}
	return exsistBalInterceptors, false, jobVolumeMount, jobVolume, errBalInterceptor, nil
}
