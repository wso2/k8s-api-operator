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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/cbroglie/mustache"
	"github.com/heroku/docker-registry-client/registry"

	wso2v1alpha1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/wso2/v1alpha1"
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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/wso2/k8s-apim-operator/apim-operator/pkg/controller/ratelimiting"
)

var log = logf.Log.WithName("controller_api")

//XMGWProductionEndpoints represents the structure of endpoint
type XMGWProductionEndpoints struct {
	Urls []string `yaml:"urls" json:"urls"`
}

//This struct use to import multiple certificates to trsutstore
type DockerfileArtifacts struct {
	CertFound    bool
	Password     string
	Certs        map[string]string
	BaseImage    string
	RuntimeImage string
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
		return reconcile.Result{}, ownerErr
	}
	userNameSpace := instance.Namespace

	//get configurations file for the controller
	controlConf, err := getConfigmap(r, controllerConfName, wso2NameSpaceConst)
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
	mgwToolkitImg := controlConfigData[mgwToolkitImgConst]
	mgwRuntimeImg := controlConfigData[mgwRuntimeImgConst]
	kanikoImg := controlConfigData[kanikoImgConst]
	dockerRegistry := controlConfigData[dockerRegistryConst]
	reqLogger.Info("Controller Configurations", "mgwToolkitImg", mgwToolkitImg, "mgwRuntimeImg", mgwRuntimeImg,
		"kanikoImg", kanikoImg, "dockerRegistry", dockerRegistry, "userNameSpace", userNameSpace)

	//creates the docker configs in the required format
	dockerSecretEr := dockerConfigCreator(r, operatorOwner, userNameSpace)
	if dockerSecretEr != nil {
		log.Error(dockerSecretEr, "Error in docker-config creation")
	}

	//Handles policy.yaml.
	//If there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies
	policyEr := policyHandler(r, operatorOwner, userNameSpace)
	if policyEr != nil {
		log.Error(policyEr, "Error in default policy map creation")
	}

	//Check if the configmap mentioned in crd object exist
	apiConfigMapRef := instance.Spec.Definition.ConfigmapName
	apiConfigMap, err := getConfigmap(r, apiConfigMapRef, userNameSpace)
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
	apiConfigMap.OwnerReferences = owner
	errorUpdateConf := r.client.Update(context.TODO(), apiConfigMap)
	if errorUpdateConf != nil {
		log.Error(errorUpdateConf, "error in updating swagger config map with owner reference")
	}
	//Fetch swagger data from configmap, reads and modifies swagger
	swaggerDataMap := apiConfigMap.Data
	swagger, swaggerDataFile, err := mgwSwaggerLoader(swaggerDataMap)

	modeExt, isModeDefined := swagger.Extensions[deploymentMode]
	mode := privateJet
	if isModeDefined {
		modeRawStr, _ := modeExt.(json.RawMessage)
		err = json.Unmarshal(modeRawStr, &mode)
		if err != nil {
			log.Info("Error unmarshal data of mode")
		}
	} else {
		log.Info("Deployment mode is not set in the swagger. Hence default to privateJet mode")
	}

	image := strings.ToLower(strings.ReplaceAll(swagger.Info.Title, " ", ""))
	tag := swagger.Info.Version
	if instance.Spec.UpdateTimeStamp != "" {
		tag = tag + "-" + instance.Spec.UpdateTimeStamp
	}
	imageName := image + ":" + tag
	// check if the image already exists
	imageExist, errImage := isImageExist(dockerRegistry+"/"+image, tag, r)
	if errImage != nil {
		log.Error(errImage, "Error in image finding")
	}
	log.Info("image exist? " + strconv.FormatBool(imageExist))
	endpointNames, newSwagger := mgwSwaggerHandler(r, swagger, mode, userNameSpace, instance.Name)
	for endpointNameL, _ := range endpointNames {
		log.Info("Endpoint name " + endpointNameL)
	}
	var containerList []corev1.Container
	//Creating sidecar endpoint deployment
	if mode == sidecar {
		for endpointName, _ := range endpointNames {
			if endpointName != "" {
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				erCr := r.client.Get(context.TODO(),
					types.NamespacedName{Namespace: userNameSpace, Name: endpointName}, targetEndpointCr)
				if erCr == nil {
					if targetEndpointCr.Spec.Deploy.DockerImage != "" {
						sidecarContainer := corev1.Container{
							Image: targetEndpointCr.Spec.Deploy.DockerImage,
							Name:  targetEndpointCr.Spec.Deploy.Name,
							Ports: []corev1.ContainerPort{{
								ContainerPort: targetEndpointCr.Spec.Port,
							}},
						}
						containerList = append(containerList, sidecarContainer)
						if err := r.reconcileSidecarEndpointService(targetEndpointCr, userNameSpace, instance); err != nil {
							return reconcile.Result{}, err
						}
					}
				} else {
					log.Info("Failed to deploy the sidecar endpoint " + endpointName)
					return reconcile.Result{}, erCr
				}
			}
		}
	}

	reqLogger.Info("getting security instance")

	var alias string
	//keep to track the existance of certificates
	var existcert bool
	//to add multiple certs with alias
	certList := make(map[string]string)
	var certName string
	//get the volume mounts
	jobVolumeMount, jobVolume := getVolumes(instance)
	//get all the securities defined in swagger
	var securityMap = make(map[string][]string)
	var securityDefinition = make(map[string]securitySchemeStruct)
	//check security scheme already exist
	_, secSchemeDefined := swagger.Extensions[securitySchemeExtension]
	//get security instances
	//get API level security
	apiLevelSecurity, isDefined := swagger.Extensions[securityExtension]
	var APILevelSecurity []map[string][]string
	if isDefined {
		log.Info("API level security is defined")
		rawmsg := apiLevelSecurity.(json.RawMessage)
		errsec := json.Unmarshal(rawmsg, &APILevelSecurity)
		if errsec != nil {
			log.Error(err, "error unmarshaling API level security ")
			return reconcile.Result{}, errsec
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
	var securityDef path
	if resSecIsDefined {
		rawmsg := resLevelSecurity.(json.RawMessage)
		errrSec := json.Unmarshal(rawmsg, &resSecurityMap)
		if errrSec != nil {
			log.Error(errrSec, "error unmarshall into resource level security")
			return reconcile.Result{}, err
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
	securityInstance := &wso2v1alpha1.Security{}
	var certificateSecret = &corev1.Secret{}
	for secName, scopeList := range securityMap {
		//retrieve security instances
		errGetSec := r.client.Get(context.TODO(), types.NamespacedName{Name: secName, Namespace: userNameSpace}, securityInstance)
		if errGetSec != nil && errors.IsNotFound(errGetSec) {
			reqLogger.Info("defined security instance " + secName + " is not found")
			return reconcile.Result{}, errGetSec
		}
		//get certificate for JWT and Oauth
		if strings.EqualFold(securityInstance.Spec.Type, securityOauth) || strings.EqualFold(securityInstance.Spec.Type, securityJWT) {
			errc := r.client.Get(context.TODO(), types.NamespacedName{Name: securityInstance.Spec.Certificate, Namespace: userNameSpace}, certificateSecret)
			if errc != nil && errors.IsNotFound(errc) {
				reqLogger.Info("defined certificate is not found")
				return reconcile.Result{}, errc
			} else {
				log.Info("defined certificate successfully retrieved")
			}
			//mount certs
			volumemountTemp, volumeTemp := certMoutHandler(r, certificateSecret, jobVolumeMount, jobVolume)
			jobVolumeMount = volumemountTemp
			jobVolume = volumeTemp
			alias = certificateSecret.Name + certAlias
			existcert = true
			for k := range certificateSecret.Data {
				certName = k
			}
			//add cert path and alias as key value pairs
			certList[alias] = certPath + certificateSecret.Name + "/" + certName
		}
		if strings.EqualFold(securityInstance.Spec.Type, securityOauth) {
			//get the keymanager server URL from the security kind
			keymanagerServerurl = securityInstance.Spec.Endpoint
			//fetch credentials from the secret created
			errGetCredentials := getCredentials(r, securityInstance.Spec.Credentials, securityOauth, userNameSpace)
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
		if strings.EqualFold(securityInstance.Spec.Type, securityJWT) {
			log.Info("retrieving data for security type JWT")
			certificateAlias = alias
			if securityInstance.Spec.Issuer != "" {
				issuer = securityInstance.Spec.Issuer
			}
			if securityInstance.Spec.Audience != "" {
				audience = securityInstance.Spec.Audience
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, basicSecurityAndScheme) {
			existcert = false
			//fetch credentials from the secret created
			errGetCredentials := getCredentials(r, securityInstance.Spec.Credentials, "Basic", userNameSpace)
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
	//create configmap with modified swagger
	swaggerConfMap := createConfigMap(apiConfigMapRef + "-mgw", swaggerDataFile, formattedSwagger, userNameSpace, owner)
	log.Info("Creating swagger configmap for mgw")
	_, errgetConf := getConfigmap(r, apiConfigMapRef + "-mgw", userNameSpace)
	if errgetConf != nil && errors.IsNotFound(errgetConf){
		log.Info("swagger-mgw is not found. Hence creating new configmap")
		errConf := r.client.Create(context.TODO(), swaggerConfMap)
		if errConf != nil {
			log.Error(err, "Error in mgw swagger configmap create")
		}
	} else if errgetConf != nil {
		log.Error(errgetConf,"error getting swagger-mgw")
	}

	if isDefined == false && len(securityDef.Security) == 0 {
		log.Info("use default security")
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
			errCertUserns := r.client.Get(context.TODO(), types.NamespacedName{Name: securityDefault.Spec.Certificate, Namespace: userNameSpace}, defaultCert)
			if errCertUserns != nil && errors.IsNotFound(errCertUserns) {
				log.Info("default certificate is not found in " + userNameSpace + "namespace")
				log.Info("retrieve default certificate from " + wso2NameSpaceConst)
				var defaultCertName string
				var defaultCertvalue []byte
				errc := r.client.Get(context.TODO(), types.NamespacedName{Name: securityDefault.Spec.Certificate, Namespace: wso2NameSpaceConst}, defaultCert)
				if errc != nil && errors.IsNotFound(errc) {
					reqLogger.Info("defined certificate is not found in " + wso2NameSpaceConst)
					return reconcile.Result{}, errc
				} else if errc != nil {
					log.Error(errc, "error in getting default cert from "+wso2NameSpaceConst)
					return reconcile.Result{}, errc
				}
				//copying default cert as a secret to user namespace
				noOwner := []metav1.OwnerReference{}
				for cert, value := range defaultCert.Data {
					defaultCertName = cert
					defaultCertvalue = value
				}
				newDefaultSecret := createSecret(securityDefault.Spec.Certificate, defaultCertName, string(defaultCertvalue), userNameSpace, noOwner)
				errCreateSec := r.client.Create(context.TODO(), newDefaultSecret)
				if errCreateSec != nil {
					log.Error(errCreateSec, "error creating secret for default security in user namespace")
					return reconcile.Result{}, errCreateSec
				}
			} else if errCertUserns != nil {
				log.Error(errCertUserns, "error in getting default certificate from "+userNameSpace+"namespace")
				return reconcile.Result{}, errCertUserns
			}
			//copying default security to user namespace
			log.Info("copying default security to " + userNameSpace)
			newDefaultSecurity := copyDefaultSecurity(securityDefault, userNameSpace)
			errCreateSecurity := r.client.Create(context.TODO(), newDefaultSecurity)
			if errCreateSecurity != nil {
				log.Error(errCreateSecurity, "error creating secret for default security in user namespace")
				return reconcile.Result{}, errCreateSecurity
			}
			log.Info("default security successfully copied to " + userNameSpace + " namespace")
		} else if errGetSec != nil {
			log.Error(errGetSec, "error getting default security from user namespace")
			return reconcile.Result{}, errGetSec
		} else {
			log.Info("default security exists in " + userNameSpace + " namespace")
		}
	}
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
					existcert = true
					analyticsEnabled = "true"
					certList[analyticsAlias] = analyticsCertLocation + fileName
				}
			}
		}
	}

	//Handles the creation of dockerfile configmap
	dockerfileConfmap, errDocker := dockerfileHandler(r, certList, existcert, controlConfigData, owner, instance)
	if errDocker != nil {
		log.Error(errDocker, "error in docker configmap handling")
		return reconcile.Result{}, errDocker
	} else {
		log.Info("kaniko job related dockerfile was written into configmap " + dockerfileConfmap.Name)
	}

	//Get data from apim configmap
	apimConfig, apimEr := getConfigmap(r, apimConfName, wso2NameSpaceConst)
	if apimEr == nil {
		verifyHostname = apimConfig.Data[verifyHostnameConst]
	} else {
		verifyHostname = verifyHostNameVal
	}

	//writes into the conf file
	filename := mgwConfTemplatePath
	output, err := mustache.RenderFile(filename, map[string]string{
		keystorePathConst:                   keystorePath,
		keystorePasswordConst:               keystorePassword,
		truststorePathConst:                 truststorePath,
		truststorePasswordConst:             truststorePassword,
		keymanagerServerurlConst:            keymanagerServerurl,
		keymanagerUsernameConst:             keymanagerUsername,
		keymanagerPasswordConst:             keymanagerPassword,
		issuerConst:                         issuer,
		audienceConst:                       audience,
		certificateAliasConst:               certificateAlias,
		enabledGlobalTMEventPublishingConst: enabledGlobalTMEventPublishing,
		basicUsernameConst:                  basicUsername,
		basicPasswordConst:                  basicPassword,
		analyticsEnabledConst:               analyticsEnabled,
		analyticsUsernameConst:              analyticsUsername,
		analyticsPasswordConst:              analyticsPassword,
		uploadingTimeSpanInMillisConst:      uploadingTimeSpanInMillis,
		rotatingPeriodConst:                 rotatingPeriod,
		uploadFilesConst:                    uploadFiles,
		verifyHostnameConst:                 verifyHostname,
		hostnameConst:                       hostname,
		portConst:                           port})

	if err != nil {
		log.Error(err, "error in rendering ")
	}
	//writes the created conf file to secret
	errCreateSecret := createMGWSecret(r, output, owner, instance)
	if errCreateSecret != nil {
		log.Error(errCreateSecret, "Error in creating conf secret")
	} else {
		log.Info("Successfully created secret")
	}

	generateK8sArtifacsForMgw := controlConfigData[generatekubernbetesartifactsformgw]
	genArtifacts, errGenArtifacts := strconv.ParseBool(generateK8sArtifacsForMgw)
	if errGenArtifacts != nil {
		log.Error(errGenArtifacts, "error reading value for generate k8s artifacts")
	}
	getResourceReqCPU := controlConfigData[resourceRequestCPU]
	getResourceReqMemory := controlConfigData[resourceRequestMemory]
	getResourceLimitCPU := controlConfigData[resourceLimitCPU]
	getResourceLimitMemory := controlConfigData[resourceLimitMemory]

	analyticsEnabledBool, _ := strconv.ParseBool(analyticsEnabled)
	dep := createMgwDeployment(instance, imageName, controlConf, analyticsEnabledBool, r, userNameSpace, owner,
		getResourceReqCPU, getResourceReqMemory, getResourceLimitCPU, getResourceLimitMemory, containerList)
	depFound := &appsv1.Deployment{}
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)

	svc := createMgwLBService(instance, userNameSpace, owner)
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
	GetAvgUtilMemory := controlConfigData[hpaTargetAverageUtilizationMemory]
	intValueUtilMemory, err := strconv.ParseInt(GetAvgUtilMemory, 10, 32)
	if err != nil {
		log.Error(err, "error getting hpa target average utilization for memory")
	}
	targetAvgUtilizationCPU := int32(intValueUtilCPU)
	targetAvgUtilizationMemory := int32(intValueUtilMemory)
	minReplicas := int32(instance.Spec.Replicas)
	errGettingHpa := createHorizontalPodAutoscaler(dep, r, owner, minReplicas, maxReplicas, targetAvgUtilizationCPU,
		targetAvgUtilizationMemory)
	if errGettingHpa != nil {
		log.Error(errGettingHpa, "Error getting HPA")
	}

	if instance.Spec.UpdateTimeStamp != "" {
		//Schedule Kaniko pod
		reqLogger.Info("Updating the API", "API.Name", instance.Name, "API.Namespace", instance.Namespace)
		job := scheduleKanikoJob(instance, imageName, controlConf, jobVolumeMount, jobVolume, instance.Spec.UpdateTimeStamp, owner)
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
		errDeleteJob := deleteCompletedJobs(instance.Namespace)
		if errDeleteJob != nil {
			log.Error(errDeleteJob, "error deleting completed jobs")
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
				return reconcile.Result{}, nil
			} else {
				log.Info("skip updating kubernetes artifacts")
				return reconcile.Result{}, nil
			}
		} else {
			reqLogger.Info("Job is still not completed.", "Job.Status", job.Status)
			return reconcile.Result{Requeue: true}, nil
		}

	} else if imageExist {
		log.Info("Image already exist, hence skipping the kaniko job")
		errDeleteJob := deleteCompletedJobs(instance.Namespace)
		if errDeleteJob != nil {
			log.Error(errDeleteJob, "error deleting completed jobs")
		}

		if genArtifacts {
			log.Info("generating kubernetes artifacts")
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

			if svcErr != nil && errors.IsNotFound(svcErr) {
				reqLogger.Info("Creating a new Service", "SVC.Namespace", svc.Namespace, "SVC.Name", svc.Name)
				svcErr = r.client.Create(context.TODO(), svc)
				if svcErr != nil {
					return reconcile.Result{}, svcErr
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
		job := scheduleKanikoJob(instance, imageName, controlConf, jobVolumeMount, jobVolume, instance.Spec.UpdateTimeStamp, owner)
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
	minReplicas int32, maxReplicas int32, targetAverageUtilizationCPU int32, targetAverageUtilizationMemory int32) error {

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
	return swagger, swaggerDataFile, err
}

//Get endpoint from swagger and replace it with targetendpoint kind service endpoint
func mgwSwaggerHandler(r *ReconcileAPI, swagger *openapi3.Swagger, mode string, userNameSpace string, apiName string) (map[string]string, *openapi3.Swagger) {

	var editedSwaggerData string
	var mgwSwagger *openapi3.Swagger
	var errMgwSwgr error
	//var resLevelEp = make(map[*openapi3.PathItem]XMGWProductionEndpoints)
	var resLevelEp = make(map[string]XMGWProductionEndpoints)
	mapName := apiName + "-swagger-mgw"
	//get mgw swagger if available
	configmapOfNewSwagger, err := getConfigmap(r, mapName,userNameSpace)
	if err != nil && errors.IsNotFound(err){
		log.Info("configmap for mgw swagger is not found.Creating a new configmap")
		mgwSwagger = swagger
	} else if err != nil {
		log.Error(err,"error getting configmap of mgw swagger file")
	}else {
		for _, value := range configmapOfNewSwagger.Data {
			editedSwaggerData = value
		}
		mgwSwagger, errMgwSwgr = openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(editedSwaggerData))
		if errMgwSwgr != nil {
			log.Error(errMgwSwgr, "error generate swagger for mgw")
		}
	}
	endpointNames := make(map[string]string)
	var checkt []string
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
				currentService := &corev1.Service{}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
					Name: endPoint}, currentService)
				erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: endPoint}, targetEndpointCr)

				if err != nil && errors.IsNotFound(err) && mode != sidecar {
					log.Error(err, "Service is not found")
				} else if erCr != nil && errors.IsNotFound(erCr) {
					log.Error(err, "targetEndpoint CRD object is not found")
				} else if err != nil && mode != sidecar {
					log.Error(err, "Error in getting service")
				} else if erCr != nil {
					log.Error(err, "Error in getting targetendpoint CRD object")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					//endpointNames[endPoint] = endPoint
					if mode == sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						endpointNames[targetEndpointCr.Name] = endPointSidecar
						checkt = append(checkt, endPointSidecar)
					} else {
						endPoint = protocol + "://" + endPoint
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
								urlVal = protocol + "://" + urlVal
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

	//resource level endpoint
	for pathName, path := range swagger.Paths {
		var checkr []string
		resourceEndpointData, checkResourceEP := path.Get.Extensions[endpointExtension]
		if checkResourceEP {
			prodEp := XMGWProductionEndpoints{}
			var endPoint string
			ResourceEndpointJson, checkJsonResource := resourceEndpointData.(json.RawMessage)
			if checkJsonResource {
				err := json.Unmarshal(ResourceEndpointJson, &endPoint)
				if err == nil {
					//check if service & targetendpoint cr object are available
					currentService := &corev1.Service{}
					targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
					err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace,
						Name: endPoint}, currentService)
					erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: userNameSpace, Name: endPoint}, targetEndpointCr)
					if err != nil && errors.IsNotFound(err) && mode != sidecar{
						log.Error(err, "Service is not found")
					} else if erCr != nil && errors.IsNotFound(erCr) {
						log.Error(err, "targetendpoint CRD object is not found")
					} else if err != nil && mode != sidecar{
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
							endPoint = protocol + "://" + endPoint
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
										urlVal = protocol + "://" + urlVal
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
	}
	for pathName, path := range mgwSwagger.Paths {
		for mapPath, value := range resLevelEp{
			if strings.EqualFold(pathName,mapPath) {
				path.Get.Extensions[endpointExtension] = value
			}
		}
	}
	return endpointNames, mgwSwagger
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
		basicPassword = strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	}
	if securityType == "Oauth" {
		keymanagerUsername = usrname
		keymanagerPassword = string(password)
	}
	return nil
}

// generate relevant MGW deployment/services for the given API definition
func createMgwDeployment(cr *wso2v1alpha1.API, imageName string, conf *corev1.ConfigMap, analyticsEnabled bool,
	r *ReconcileAPI, nameSpace string, owner []metav1.OwnerReference, resourceReqCPU string, resourceReqMemory string,
	resourceLimitCPU string, resourceLimitMemory string, containerList []corev1.Container) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	controlConfigData := conf.Data
	dockerRegistry := controlConfigData[dockerRegistryConst]
	reps := int32(cr.Spec.Replicas)
	deployVolumeMount := []corev1.VolumeMount{}
	deployVolume := []corev1.Volume{}
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
		Image:           dockerRegistry + "/" + imageName,
		ImagePullPolicy: "Always",
		Resources: corev1.ResourceRequirements{
			Requests: req,
			Limits:   lim,
		},
		VolumeMounts: deployVolumeMount,
		Ports: []corev1.ContainerPort{{
			ContainerPort: 9095,
		}},
	}

	containerList = append(containerList, apiContainer)
	return &appsv1.Deployment{
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
					Containers: containerList,
					Volumes:    deployVolume,
				},
			},
		},
	}
}

//Handles dockerfile configmap creation
func dockerfileHandler(r *ReconcileAPI, certList map[string]string, existcert bool, conf map[string]string,
	owner []metav1.OwnerReference, cr *wso2v1alpha1.API) (*corev1.ConfigMap, error) {
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
		CertFound:    existcert,
		Password:     truststorePass,
		Certs:        certList,
		BaseImage:    conf[mgwToolkitImgConst],
		RuntimeImage: conf[mgwRuntimeImgConst],
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
	if err != nil && errors.IsNotFound(err) {
		dockerConf := createConfigMap(cr.Name+"-"+dockerFile, "Dockerfile", builder.String(), cr.Namespace, owner)

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

// checks if the image exist in dockerhub
func isImageExist(image string, tag string, r *ReconcileAPI) (bool, error) {
	url := dockerhubRegistryUrl
	username := ""
	password := ""

	//checks if docker secret is available
	dockerSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dockerSecretNameConst, Namespace: wso2NameSpaceConst}, dockerSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Docker Secret is not found")
	} else if err != nil {
		log.Error(err, "error ")
	} else {
		dockerData := dockerSecret.Data
		username = string(dockerData[usernameConst])
		password = string(dockerData[passwordConst])
	}

	hub, err := registry.New(url, username, password)
	if err != nil {
		log.Error(err, "error connecting to hub")
		return false, err
	}
	tags, err := hub.Tags(image)
	if err != nil {
		log.Error(err, "error getting tags")
		return false, err
	}
	for _, foundTag := range tags {
		if foundTag == tag {
			log.Info("found the image tag")
			return true, nil
		}
	}
	return false, nil
}

//Schedule Kaniko Job to generate micro-gw image
func scheduleKanikoJob(cr *wso2v1alpha1.API, imageName string, conf *corev1.ConfigMap, jobVolumeMount []corev1.VolumeMount,
	jobVolume []corev1.Volume, timeStamp string, owner []metav1.OwnerReference) *batchv1.Job {
	//labels := map[string]string{
	//	"app": cr.Name,
	//}
	kanikoJobName := cr.Name + "kaniko"
	if timeStamp != "" {
		kanikoJobName = kanikoJobName + "-" + timeStamp
	}
	controlConfigData := conf.Data
	kanikoImg := controlConfigData[kanikoImgConst]
	dockerRegistry := controlConfigData[dockerRegistryConst]

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
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         cr.Name + "gen-container",
							Image:        kanikoImg,
							VolumeMounts: jobVolumeMount,
							Args: []string{
								"--dockerfile=/usr/wso2/dockerfile/Dockerfile",
								"--context=/usr/wso2/",
								"--destination=" + dockerRegistry + "/" + imageName,
							},
						},
					},
					RestartPolicy: "Never",
					Volumes:       jobVolume,
				},
			},
		},
	}
}

func dockerConfigCreator(r *ReconcileAPI, operatorOwner []metav1.OwnerReference, namespace string) error {
	//checks if docker secret is available
	dockerSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dockerSecretNameConst, Namespace: wso2NameSpaceConst}, dockerSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Docker Secret is not found")
		return err
	} else if err != nil {
		log.Error(err, "error ")
		return err
	}
	dockerData := dockerSecret.Data
	dockerUsername := string(dockerData[usernameConst])
	dockerPassword := string(dockerData[passwordConst])
	rawCredentials := dockerUsername + ":" + dockerPassword
	credentials := b64.StdEncoding.EncodeToString([]byte(rawCredentials))

	//make the docker-config template
	filename := "/usr/local/bin/dockerSecretTemplate.mustache"
	output, err := mustache.RenderFile(filename, map[string]string{
		"docker_url":  "https://index.docker.io/v1/",
		"credentials": credentials})
	if err != nil {
		log.Error(err, "error in rendering ")
		return err
	}

	//Writes the created template to a secret
	dockerConf := createSecret(dockerConfig, "config.json", output, namespace, operatorOwner)

	// Check if this configmap already exists
	foundsecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dockerConf.Name, Namespace: dockerConf.Namespace}, foundsecret)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new docker-config secret", "Namespace", dockerConf.Namespace, "secret.Name", dockerConf.Name)
		err = r.client.Create(context.TODO(), dockerConf)
		if err != nil {
			log.Error(err, "error ")
			return err
		}
		// secret created successfully
		return nil
	} else if err != nil {
		log.Error(err, "error ")
		return err
	}
	log.Info("Docker config secret already exists", "secret.Namespace", foundsecret.Namespace, "secret.Name", foundsecret.Name)
	log.Info("Updating Config map", "confmap.Namespace", dockerConf.Namespace, "confmap.Name", dockerConf.Name)
	err = r.client.Update(context.TODO(), dockerConf)
	if err != nil {
		log.Error(err, "error ")
		return err
	}
	return nil
}

//Service of the API
//todo: This has to be changed to LB type
func createMgwService(cr *wso2v1alpha1.API, nameSpace string) *corev1.Service {

	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: nameSpace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: "NodePort",
			Ports: []corev1.ServicePort{{
				Name:       "https",
				Protocol:   corev1.ProtocolTCP,
				Port:       9095,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 9095},
				NodePort:   30010,
			}},
			Selector: labels,
		},
	}
}

//Creating a LB balancer service to expose mgw
func createMgwLBService(cr *wso2v1alpha1.API, nameSpace string, owner []metav1.OwnerReference) *corev1.Service {

	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.Name,
			Namespace:       nameSpace,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: corev1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []corev1.ServicePort{{
				Name:       "port-9095",
				Port:       9095,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 9095},
			}, {
				Name:       "port-9090",
				Port:       9090,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 9090},
			}},
			Selector: labels,
		},
	}
}

//default volume mounts for the kaniko job
func getVolumes(cr *wso2v1alpha1.API) ([]corev1.VolumeMount, []corev1.Volume) {

	jobVolumeMount := []corev1.VolumeMount{
		{
			Name:      swaggerVolume,
			MountPath: swaggerLocation,
			ReadOnly:  true,
		},
		{
			Name:      mgwDockerFile,
			MountPath: dockerFileLocation,
		},
		{
			Name:      dockerConfig,
			MountPath: dockerConfLocation,
			ReadOnly:  true,
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

	jobVolume := []corev1.Volume{
		{
			Name: swaggerVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Spec.Definition.ConfigmapName + "-mgw",
					},
				},
			},
		},
		{
			Name: dockerConfig,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: dockerConfig,
				},
			},
		},
		{
			Name: mgwDockerFile,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Name + "-" + dockerFile,
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
					SecretName: cr.Name + "-" + mgwConfSecretConst,
				},
			},
		},
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
	jobVolumeMount = append(jobVolumeMount, corev1.VolumeMount{
		Name:      certConfig,
		MountPath: certPath + cert.Name,
		ReadOnly:  true,
	})

	jobVolume = append(jobVolume, corev1.Volume{
		Name: certConfig,
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
	//}
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
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: "apim-operator", Namespace: wso2NameSpaceConst}, depFound)
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

func copyDefaultSecurity(securityDefault *wso2v1alpha1.Security, userNameSpace string) *wso2v1alpha1.Security {

	return &wso2v1alpha1.Security{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSecurity,
			Namespace: userNameSpace,
		},
		Spec: wso2v1alpha1.SecuritySpec{
			Type:        securityDefault.Spec.Type,
			Certificate: securityDefault.Spec.Certificate,
			Audience:    securityDefault.Spec.Audience,
			Issuer:      securityDefault.Spec.Issuer,
		},
	}
}

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileAPI) createDeploymentForSidecarBackend(m *wso2v1alpha1.TargetEndpoint,
	namespace string, instance *wso2v1alpha1.API) *appsv1.Deployment {
	replicas := m.Spec.Deploy.Count
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: namespace,
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
					}},
				},
			},
		},
	}
	// Set Examplekind instance as the owner and controller
	controllerutil.SetControllerReference(instance, dep, r.scheme)
	return dep

}

func (r *ReconcileAPI) reconcileSidecarEndpointService(m *wso2v1alpha1.TargetEndpoint, namespace string,
	instance *wso2v1alpha1.API) error {
	newService := r.createServiceForSidecarEndpoint(m, namespace, instance)

	err := r.client.Create(context.TODO(), newService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Service resource: %v", err)
	}

	if err == nil {
		return nil
	}

	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: namespace,
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

// NewService assembles the ClusterIP service for the Nginx
func (r *ReconcileAPI) createServiceForSidecarEndpoint(m *wso2v1alpha1.TargetEndpoint,
	namespace string, instance *wso2v1alpha1.API) *corev1.Service {
	var port int
	port = int(m.Spec.Port)
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ObjectMeta.Name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: m.ObjectMeta.Labels,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{Port: m.Spec.Port, TargetPort: intstr.FromInt(port)},
			},
		},
	}
	controllerutil.SetControllerReference(instance, &service, r.scheme)
	return &service
}

func (r *ReconcileAPI) reconcileDeploymentForSidecarEndpoint(m *wso2v1alpha1.TargetEndpoint, namespace string,
	instance *wso2v1alpha1.API) error {
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.createDeploymentForSidecarBackend(m, namespace, instance)
		log.WithValues("Creating a new Deployment %s/%s\n", namespace, dep.Name)
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

func deleteCompletedJobs(namespace string) error {
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
