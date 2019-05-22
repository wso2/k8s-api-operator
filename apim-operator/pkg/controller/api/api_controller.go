package api

import (
	"context"

	"github.com/cbroglie/mustache"
	wso2v1alpha1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/wso2/v1alpha1"

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
	"fmt"
	"strings"

	//"gopkg.in/yaml.v2"

	"github.com/getkin/kin-openapi/openapi3"
)

var log = logf.Log.WithName("controller_api")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
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

	//Check if the configmap mentioned in crd object exist
	apiConfigMapRef := instance.Spec.Definition.ConfigMapKeyRef.Name
	log.Info(apiConfigMapRef)

	apiConfigMap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: apiConfigMapRef, Namespace: "default"}, apiConfigMap)

	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Specified configmap is not found: %s", apiConfigMapRef)
		return reconcile.Result{}, err
	} else if err != nil {
		log.Error(err, "error ")
		return reconcile.Result{}, err
	}

	//Fetch swagger data from configmap
	swaggerDataMap := apiConfigMap.Data
	var swaggerData string
	var swaggerDataType string
	for key, value := range swaggerDataMap {
		swaggerData = value
		swaggerTemp := strings.Split(key, ".")
		swaggerDataType = swaggerTemp[1]
	}
	fmt.Println("swagger data type : ", swaggerDataType)
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(swaggerData))
	if err != nil {
		log.Error(err, "Swagger loading error ")
	}

	//get endpoint from swagger and replace it with targetendpoint kind service endpoint
	data, ok := swagger.Extensions["x-mgw-production-endpoints"]
	if ok {
		prodEp := XMGWProductionEndpoints{}
		serviceEp := ServiceEndpoints{}
		var endPoint string
		datax, ok1 := data.(json.RawMessage)
		fmt.Println(ok1)
		if ok1 {
			err = json.Unmarshal(datax, &serviceEp)

			if err == nil {
				endPoint = "https://" + serviceEp.ServiceName
				checkt := []string{endPoint}
				prodEp.Urls = checkt

				currentService := &corev1.Service{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: "default",
					Name: serviceEp.ServiceName}, currentService)

				if err != nil && errors.IsNotFound(err) {
					log.Error(err, "Service CRD object is not found")
				} else if err != nil {
					log.Error(err, "Error in getting service")
				} else {
					swagger.Extensions["x-mgw-production-endpoints"] = prodEp
				}

			}
		}
	}

	//get throttling tiers and replace with ratelimiting kind


	final, err := swagger.MarshalJSON()
	fmt.Println(string(final))

	// gets the data from analytics secret
	analyticsData, err := getSecretData(r)

	//writes into the conf file

	if err == nil && analyticsData != nil && analyticsData["username"] != nil &&
		analyticsData["password"] != nil {
		analyticsEnabled = "true"
		analyticsUsername = string(analyticsData["username"])
		analyticsPassword = string(analyticsData["password"])
	}

	filename := "/usr/local/bin/microgwconf.mustache"
	output, err := mustache.RenderFile(filename, map[string]string{
		"keystorePath":                   keystorePath,
		"keystorePassword":               keystorePassword,
		"truststorePath":                 truststorePath,
		"truststorePassword":             truststorePassword,
		"keymanagerServerurl":            keymanagerServerurl,
		"keymanagerUsername":             keymanagerUsername,
		"keymanagerPassword":             keymanagerPassword,
		"issuer":                         issuer,
		"audience":                       audience,
		"certificateAlias":               certificateAlias,
		"enabledGlobalTMEventPublishing": enabledGlobalTMEventPublishing,
		"basicUsername":                  basicUsername,
		"basicPassword":                  basicPassword,
		"analyticsEnabled":               analyticsEnabled,
		"analyticsUsername":              analyticsUsername,
		"analyticsPassword":              analyticsPassword})

	//fmt.Println(output)

	if err != nil {
		log.Error(err, "error in rendering ")
	}

	//writes the created conf file to secret
	errCreateSecret := createMGWSecret(r, output)
	if errCreateSecret != nil {
		log.Error(errCreateSecret, "Error in creating conf secret")
	} else {
		log.Info("Successfully created secret")
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set API instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *wso2v1alpha1.API) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// gets the data from analytics secret
func getSecretData(r *ReconcileAPI) (map[string][]byte, error) {
	var analyticsData map[string][]byte
	// Check if this secret exists
	analyticsSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "analytics-secret", Namespace: "wso2-system"}, analyticsSecret)

	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Analytics Secret is not found")
		return analyticsData, err

	} else if err != nil {
		log.Error(err, "error ")
		return analyticsData, err

	}

	analyticsData = analyticsSecret.Data
	log.Info("Analytics Secret exists")
	fmt.Println("DATA")
	fmt.Println(string(analyticsData["username"]))
	fmt.Println(string(analyticsData["password"]))
	fmt.Println("END")
	return analyticsData, nil

}

func createMGWSecret(r *ReconcileAPI, confData string) error {
	var apimSecret *corev1.Secret

	apimSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mgw-secret",
			Namespace: "wso2-system",
		},
	}

	apimSecret.Data = map[string][]byte{
		"confData": []byte(confData),
	}

	// Check if this secret exists
	checkSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "mgw-secret", Namespace: "wso2-system"}, checkSecret)

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
