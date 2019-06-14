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
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/cbroglie/mustache"
	"github.com/heroku/docker-registry-client/registry"

	wso2v1alpha1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/wso2/v1alpha1"
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
type Certs struct {
	CertFound bool
	Password  string
	Certs     map[string]string
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

	//get configurations file for the controller
	controlConf, err := getConfigmap(r, controllerConfName, wso2NameSpaceConst)
	if err != nil {
		if errors.IsNotFound(err) {
			// Controller configmap is not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	controlConfigData := controlConf.Data
	mgwToolkitImg := controlConfigData[mgwToolkitImgConst]
	mgwRuntimeImg := controlConfigData[mgwRuntimeImgConst]
	kanikoImg := controlConfigData[kanikoImgConst]
	dockerRegistry := controlConfigData[dockerRegistryConst]
	userNameSpace := controlConfigData[userNameSpaceConst]
	reqLogger.Info("Controller Configurations", "mgwToolkitImg", mgwToolkitImg, "mgwRuntimeImg", mgwRuntimeImg,
		"kanikoImg", kanikoImg, "dockerRegistry", dockerRegistry, "userNameSpace", userNameSpace)

	dockerSecretEr := dockerConfigCreator(r)
	if dockerSecretEr != nil {
		log.Error(dockerSecretEr, "Error in docker-config creation")
	}
	//Handles policy.yaml.
	//If there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies

	policyEr := policyHandler(r)
	if policyEr != nil {
		log.Error(policyEr, "Error in default policy map creation")
	}

	//Check if the configmap mentioned in crd object exist
	apiConfigMapRef := instance.Spec.Definition.ConfigMapKeyRef.Name
	log.Info(apiConfigMapRef)
	apiConfigMap, err := getConfigmap(r, apiConfigMapRef, wso2NameSpaceConst)
	if err != nil {
		if errors.IsNotFound(err) {
			// Swagger configmap is not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Fetch swagger data from configmap, reads and modifies swagger
	swaggerDataMap := apiConfigMap.Data
	swagger, swaggerDataFile, err := mgwSwaggerLoader(swaggerDataMap)
	if err != nil {
		log.Error(err, "Swagger loading error ")
	}

	image := strings.ToLower(strings.ReplaceAll(swagger.Info.Title, " ", ""))
	tag := swagger.Info.Version
	imageName := image + ":" + tag

	// check if the image already exists
	imageExist, errImage := isImageExist(dockerRegistry+"/"+image, tag, r)
	if errImage != nil {
		log.Error(errImage, "Error in image finding")
	}
	log.Info("image exist? " + strconv.FormatBool(imageExist))

	newSwagger := mgwSwaggerHandler(r, swagger)

	//update configmap with modified swagger

	swaggerConfMap, err := createConfigMap(apiConfigMapRef, swaggerDataFile, newSwagger, wso2NameSpaceConst)
	if err != nil {
		log.Error(err, "Error in modified swagger configmap structure")
	}

	log.Info("Updating swagger configmap")
	errConf := r.client.Update(context.TODO(), swaggerConfMap)
	if errConf != nil {
		log.Error(err, "Error in modified swagger configmap update")
	}

	reqLogger.Info("getting security instance")

	//get defined security cr from swagger
	definedSecurity, checkSecuritykind := swagger.Extensions[securityExtension]
	var securityName string

	if checkSecuritykind {
		rawmsg := definedSecurity.(json.RawMessage)
		errsec := json.Unmarshal(rawmsg, &securityName)

		if errsec != nil {
			log.Error(err, "error getting security kind from swagger ")
			return reconcile.Result{}, errsec
		}

	} else {
		//use default security cr
		securityName = defaultSecurity
	}
	//get security instance. sample secret name is hard coded for now.
	security := &wso2v1alpha1.Security{}
	errGetSec := r.client.Get(context.TODO(), types.NamespacedName{Name: securityName, Namespace: wso2NameSpaceConst}, security)

	if errGetSec != nil && errors.IsNotFound(errGetSec) {
		reqLogger.Info("defined security instance is not found")
		return reconcile.Result{}, errGetSec
	}

	var certificateSecret = &corev1.Secret{}
	var alias string
	//keep to track the existance of certificates
	var existcert bool
	//to add multiple certs with alias
	certList := make(map[string]string)
	var certName string
	//get the volume mounts
	jobVolumeMount, jobVolume := getVolumes(instance)

	//get certificate for JWT and Oauth
	if strings.EqualFold(security.Spec.Type, "Oauth") || strings.EqualFold(security.Spec.Type, "JWT") {

		errc := r.client.Get(context.TODO(), types.NamespacedName{Name: security.Spec.Certificate, Namespace: wso2NameSpaceConst}, certificateSecret)

		if errc != nil && errors.IsNotFound(errc) {
			reqLogger.Info("defined cretificate is not found")
			return reconcile.Result{}, errc
		}

		volumemountTemp, volumeTemp := certMoutHandler(r, certificateSecret, jobVolumeMount, jobVolume)
		jobVolumeMount = volumemountTemp
		jobVolume = volumeTemp

		alias = certificateSecret.Name + "alias"
		existcert = true

		for k, _ := range certificateSecret.Data {
			certName = k
		}

		//add cert path and alias as key value pairs
		certList[alias] = certPath + certificateSecret.Name + "/" + certName
	}

	if strings.EqualFold(security.Spec.Type, "Oauth") {
		//fetch credentials from the secret created
		fmt.Println("security type Oauth")
		errGetCredentials := getCredentials(r, security.Spec.Credentials, "Oauth")

		if errGetCredentials != nil {
			log.Error(errGetCredentials, "Error occured when retriving credentials for Oauth")
		} else {
			log.Info("Credentials successfully retrived")
		}
	}

	if strings.EqualFold(security.Spec.Type, "JWT") {

		certificateAlias = alias

		if security.Spec.Issuer != "" {
			issuer = security.Spec.Issuer
		}
		if security.Spec.Audience != "" {
			audience = security.Spec.Audience
		}
	}

	if strings.EqualFold(security.Spec.Type, "Basic") {
		existcert = false
		//fetch credentials from the secret created
		errGetCredentials := getCredentials(r, security.Spec.Credentials, "Basic")

		if errGetCredentials != nil {
			log.Error(errGetCredentials, "Error occured when retriving credentials for Basic")
		} else {
			log.Info("Credentials successfully retrived")
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
			verifyHostname = analyticsConf.Data[verifyHostnameConst]
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
				jobVolumeMountTemp, jobVolumeTemp, fileName, errCert := analyticsVolumeHandler(analyticsCertSecretName, r, jobVolumeMount, jobVolume)
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

	//writes into the conf file
	//Handles the creation of dockerfile configmap
	dockerfileConfmap, errDocker := dockerfileHandler(r, certList, existcert)
	if errDocker != nil {
		log.Error(errDocker, "error in docker configmap handling")
	}
	log.Info("docker file data " + dockerfileConfmap.Data["Dockerfile"])

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

	fmt.Println(output)

	//writes the created conf file to secret
	errCreateSecret := createMGWSecret(r, output)
	if errCreateSecret != nil {
		log.Error(errCreateSecret, "Error in creating conf secret")
	} else {
		log.Info("Successfully created secret")
	}

	//Schedule Kaniko pod
	job := scheduleKanikoJob(instance, imageName, controlConf, jobVolumeMount, jobVolume)
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

	analyticsEnabledBool, _ := strconv.ParseBool(analyticsEnabled)
	dep := createMgwDeployment(instance, imageName, controlConf, analyticsEnabledBool, r)
	depFound := &appsv1.Deployment{}
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)

	svc := createMgwLBService(instance, userNameSpace)
	svcFound := &corev1.Service{}
	svcErr := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, svcFound)

	// if kaniko job is succeeded, create the deployment
	if kubeJob.Status.Succeeded > 0 {
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
		reqLogger.Info("Job is still not completed.", "Job.Status", job.Status)
		return reconcile.Result{}, deperr
	}
	// Job already exists - don't requeue
	reqLogger.Info("Skip reconcile: Job already exists", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
	return reconcile.Result{}, jobErr
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

func createMGWSecret(r *ReconcileAPI, confData string) error {
	var apimSecret *corev1.Secret

	apimSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mgwConfSecretConst,
			Namespace: wso2NameSpaceConst,
		},
	}

	apimSecret.Data = map[string][]byte{
		mgwConfConst: []byte(confData),
	}

	// Check if this secret exists
	checkSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: mgwConfSecretConst, Namespace: wso2NameSpaceConst}, checkSecret)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating secret ")
		errSecret := r.client.Create(context.TODO(), apimSecret)
		return errSecret
	} else if err != nil {
		log.Error(err, "error ")
		return err
	} else {
		log.Info("Updating secret")
		errSecret := r.client.Update(context.TODO(), apimSecret)
		return errSecret
	}

}

//get configmap
func getConfigmap(r *ReconcileAPI, mapName string, ns string) (*corev1.ConfigMap, error) {
	apiConfigMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: mapName, Namespace: ns}, apiConfigMap)

	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Specified configmap is not found: %s", mapName)
		return apiConfigMap, err
	} else if err != nil {
		log.Error(err, "error ")
		return apiConfigMap, err
	} else {
		return apiConfigMap, nil
	}

}

// createConfigMap creates a config file with the given data
func createConfigMap(apiConfigMapRef string, key string, value string, ns string) (*corev1.ConfigMap, error) {

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiConfigMapRef,
			Namespace: ns,
		},
		Data: map[string]string{
			key: value,
		},
	}, nil
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
func mgwSwaggerHandler(r *ReconcileAPI, swagger *openapi3.Swagger) string {

	//api level endpoint
	endpointData, checkEndpoint := swagger.Extensions[endpointExtension]
	if checkEndpoint {
		prodEp := XMGWProductionEndpoints{}
		var endPoint string
		endpointJson, checkJsonRaw := endpointData.(json.RawMessage)

		if checkJsonRaw {
			err := json.Unmarshal(endpointJson, &endPoint)
			if err == nil {
				//check if service & targetendpoint cr object are available
				currentService := &corev1.Service{}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: wso2NameSpaceConst,
					Name: endPoint}, currentService)
				erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: wso2NameSpaceConst, Name: endPoint}, targetEndpointCr)

				if err != nil && errors.IsNotFound(err) {
					log.Error(err, "Service is not found")
				} else if erCr != nil && errors.IsNotFound(erCr) {
					log.Error(err, "targetendpoint CRD object is not found")
				} else if err != nil {
					log.Error(err, "Error in getting service")
				} else if erCr != nil {
					log.Error(err, "Error in getting targetendpoint CRD object")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					endPoint = protocol + "://" + endPoint
					checkt := []string{endPoint}
					prodEp.Urls = checkt
					swagger.Extensions[endpointExtension] = prodEp
				}
			}
		}
	}

	//resource level endpoint
	for _, path := range swagger.Paths {
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
					err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: wso2NameSpaceConst,
						Name: endPoint}, currentService)
					erCr := r.client.Get(context.TODO(), types.NamespacedName{Namespace: wso2NameSpaceConst, Name: endPoint}, targetEndpointCr)

					if err != nil && errors.IsNotFound(err) {
						log.Error(err, "Service is not found")
					} else if erCr != nil && errors.IsNotFound(erCr) {
						log.Error(err, "targetendpoint CRD object is not found")
					} else if err != nil {
						log.Error(err, "Error in getting service")
					} else if erCr != nil {
						log.Error(err, "Error in getting targetendpoint CRD object")
					} else {
						protocol := targetEndpointCr.Spec.Protocol
						endPoint = protocol + "://" + endPoint
						checkt := []string{endPoint}
						prodEp.Urls = checkt
						path.Get.Extensions[endpointExtension] = prodEp
					}
				}
			}
		}
	}

	//reformatting swagger
	var prettyJSON bytes.Buffer
	final, err := swagger.MarshalJSON()
	if err != nil {
		log.Error(err, "swagger marshal error")
	}
	errIndent := json.Indent(&prettyJSON, final, "", "  ")
	if errIndent != nil {
		log.Error(errIndent, "Error in pretty json")
	}

	newSwagger := string(prettyJSON.Bytes())

	return newSwagger

}

func getCredentials(r *ReconcileAPI, name string, securityType string) error {

	hasher := sha1.New()
	var usrname string
	var password []byte

	//get the secret included credentials
	credentialSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: wso2NameSpaceConst}, credentialSecret)

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
	r *ReconcileAPI) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}

	controlConfigData := conf.Data
	dockerRegistry := controlConfigData[dockerRegistryConst]
	nameSpace := controlConfigData[userNameSpaceConst]
	reps := int32(cr.Spec.Definition.Replicas)

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

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: nameSpace,
			Labels:    labels,
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
					Containers: []corev1.Container{
						{
							Name:         "mgw" + cr.Name,
							Image:        dockerRegistry + "/" + imageName,
							VolumeMounts: deployVolumeMount,
							Ports: []corev1.ContainerPort{{
								ContainerPort: 9095,
							}},
						},
					},
					Volumes: deployVolume,
				},
			},
		},
	}
}

//Handles dockerfile configmap creation
func dockerfileHandler(r *ReconcileAPI, certList map[string]string, existcert bool) (*corev1.ConfigMap, error) {
	truststorePass := getTruststorePassword(r)
	dockertemplate := dockertemplatepath
	certs := &Certs{
		CertFound: existcert,
		Password:  truststorePass,
		Certs:     certList,
	}
	//generate dockerfile from the template
	tmpl, err := template.ParseFiles(dockertemplate)
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

	dockerfileConfmap, err := getConfigmap(r, dockerFile, wso2NameSpaceConst)
	if err != nil && errors.IsNotFound(err) {

		dockerConf, er := createConfigMap(dockerFile, "Dockerfile", builder.String(), wso2NameSpaceConst)
		if er != nil {
			log.Error(er, "error in docker configmap creation")
			return dockerfileConfmap, er
		}
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

func policyHandler(r *ReconcileAPI) error {
	//Check if policy configmap is available
	foundmapc := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: policyConfigmap, Namespace: wso2NameSpaceConst}, foundmapc)

	if err != nil && errors.IsNotFound(err) {
		//create new map with default policies if a map is not found
		log.Info("Creating a config map with default policies", "Namespace", wso2NameSpaceConst, "Name", policyConfigmap)

		defaultval := ratelimiting.CreateDefault()

		confmap, confer := ratelimiting.CreatePolicyConfigMap(defaultval)
		if confer != nil {
			log.Error(confer, "Error in default config map structure creation")
			return confer
		}
		foundmapc = confmap
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
	jobVolume []corev1.Volume) *batchv1.Job {
	labels := map[string]string{
		"app": cr.Name,
	}

	controlConfigData := conf.Data
	kanikoImg := controlConfigData[kanikoImgConst]
	dockerRegistry := controlConfigData[dockerRegistryConst]

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "kaniko",
			Namespace: cr.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name + "-job",
					Namespace: cr.Namespace,
					Labels:    labels,
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

func dockerConfigCreator(r *ReconcileAPI) error {
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

	//Writes the created template to a configmap
	dockerConf, er := createConfigMap(dockerConfig, "config.json", output, wso2NameSpaceConst)
	if er != nil {
		log.Error(er, "error in docker-config configmap creation")
		return er
	}

	// Check if this configmap already exists
	foundmap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dockerConf.Name, Namespace: dockerConf.Namespace}, foundmap)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Config map", "Namespace", dockerConf.Namespace, "confmap.Name", dockerConf.Name)
		err = r.client.Create(context.TODO(), dockerConf)
		if err != nil {
			log.Error(err, "error ")
			return err
		}
		// confmap created successfully
		return nil
	} else if err != nil {
		log.Error(err, "error ")
		return err
	}
	log.Info("Map already exists", "confmap.Namespace", foundmap.Namespace, "confmap.Name", foundmap.Name)
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
func createMgwLBService(cr *wso2v1alpha1.API, nameSpace string) *corev1.Service {

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
			Type: "LoadBalancer",
			Ports: []corev1.ServicePort{{
				Port:       9095,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 9095},
			}},
			Selector: labels,
		},
	}
}

//default volume mounts for the kaniko job
func getVolumes(cr *wso2v1alpha1.API) ([]corev1.VolumeMount, []corev1.Volume) {

	apiConfMap := cr.Spec.Definition.ConfigMapKeyRef.Name

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
						Name: apiConfMap,
					},
				},
			},
		},
		{
			Name: dockerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dockerConfig,
					},
				},
			},
		},
		{
			Name: mgwDockerFile,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dockerFile,
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
					SecretName: mgwConfSecretConst,
				},
			},
		},
	}

	return jobVolumeMount, jobVolume

}

// Handles the mounting of analytics certificate
func analyticsVolumeHandler(analyticsCertSecretName string, r *ReconcileAPI, jobVolumeMount []corev1.VolumeMount,
	jobVolume []corev1.Volume) ([]corev1.VolumeMount, []corev1.Volume, string, error) {
	var fileName string
	analyticsCertSecret := &corev1.Secret{}
	//checks if the certificate exists
	errCert := r.client.Get(context.TODO(), types.NamespacedName{Name: analyticsCertSecretName, Namespace: wso2NameSpaceConst}, analyticsCertSecret)

	if errCert != nil {
		log.Error(errCert, "Error in getting certificate secret specified in analytics")
	} else {
		log.Info("Analytics certificate found. Mounting it to volume.")
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
	}
	return jobVolumeMount, jobVolume, fileName, errCert
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
		log.Error(errsecret, "error in creating trustsote password")
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
