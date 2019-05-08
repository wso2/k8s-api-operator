package ratelimiting

import (
	"context"

	wso2v1alpha1 "github.com/apim-crd/apim-operator/pkg/apis/wso2/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"strconv"
	"strings"

	"fmt"

	mustache "github.com/cbroglie/mustache"
)

var log = logf.Log.WithName("controller_ratelimiting")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RateLimiting Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRateLimiting{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ratelimiting-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RateLimiting
	err = c.Watch(&source.Kind{Type: &wso2v1alpha1.RateLimiting{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner RateLimiting
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wso2v1alpha1.RateLimiting{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRateLimiting{}

// ReconcileRateLimiting reconciles a RateLimiting object
type ReconcileRateLimiting struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RateLimiting object and makes changes based on the state read
// and what is in the RateLimiting.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRateLimiting) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RateLimiting")

	// Fetch the RateLimiting instance
	instance := &wso2v1alpha1.RateLimiting{}
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

	// GENERATE POLICY CODE USING CRD INSTANCE

	nameArray := strings.Split(instance.ObjectMeta.Name, "-")
	name := nameArray[0]
	log.Info(name)

	funcName := "init" + instance.Spec.Type + name + "Policy"
	log.Info(funcName)

	tierType := instance.Spec.Type + "Tier"
	log.Info(tierType)

	policyKey := instance.Spec.Type + "Key"
	log.Info(policyKey)

	unitTime := strconv.Itoa(instance.Spec.UnitTime)
	log.Info(unitTime)

	count := strconv.Itoa(instance.Spec.RequestCount.Limit)
	log.Info(count)

	stopOnQuotaReach := strconv.FormatBool(instance.Spec.StopOnQuotaReach)
	log.Info("QUOTAREACH")
	log.Info(stopOnQuotaReach)

	filename := "/usr/local/bin/policy.mustache"
	output, err := mustache.RenderFile(filename, map[string]string{"name": name, "funcName": funcName, "tierType": tierType, "policyKey": policyKey, "unitTime": unitTime, "stopOnQuotaReach": "true", "count": count})

	log.Info(output)
	fmt.Println(output)

	if err != nil {
		log.Error(err, "error in rendering ")
	}

	//CREATE CONFIG MAP

	confmap, confEr := createConfigMap(output, name, instance)
	if confEr != nil {
		log.Error(confEr, "Error in config map structure creation")
	}

	// Check if this configmap already exists
	foundmap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: confmap.Name, Namespace: confmap.Namespace}, foundmap)

	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Config map", "confmap.Namespace", confmap.Namespace, "confmap.Name", confmap.Name)
		err = r.client.Create(context.TODO(), confmap)
		if err != nil {
			log.Error(err, "error ")
			return reconcile.Result{}, err
		}

		// confmap created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		log.Error(err, "error ")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Skip reconcile: map already exists", "confmap.Namespace", foundmap.Namespace, "confmap.Name", foundmap.Name)
	return reconcile.Result{}, nil

}

// createConfigMap creates a config file with the generated code
func createConfigMap(output string, name string, cr *wso2v1alpha1.RateLimiting) (*corev1.ConfigMap, error) {

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-configmap",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"Code": output,
		},
	}, nil
}
