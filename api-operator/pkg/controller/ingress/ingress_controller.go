package ingress

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	gwclient "github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/client"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/class"
	inghandler "github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/handler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ingress")

// TODO: (renuka) operatorNamespace represents the namespace of the operator
const operatorNamespace = "wso2-system"

var (
	// successfullyHandledRequestCount represents number of requests successfully handled by the controller for the ingresses
	// managed by this controller.
	successfullyHandledRequestCount = 0
)

// Add creates a new Ingress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIngress{
		client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		recorder:   mgr.GetEventRecorderFor("ingress-controller"),
		ingHandler: &inghandler.Handler{GatewayClient: &gwclient.Http{}},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ingress-controller", mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: 1, // MaxConcurrentReconciles should be 1 for handling ingresses
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Ingress
	err = c.Watch(&source.Kind{Type: &v1beta1.Ingress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIngress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIngress{}

// ReconcileIngress reconciles a Ingress object
type ReconcileIngress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	recorder   record.EventRecorder
	ingHandler *inghandler.Handler
}

// Reconcile reads that state of the cluster for a Ingress object and makes changes based on the state read
// and what is in the Ingress.Spec
func (r *ReconcileIngress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling Ingress")

	// Fetch the Ingress instance
	instance := &v1beta1.Ingress{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if !errors.IsNotFound(err) {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
		// Ingress object not found, could have been deleted after reconcile request.
		// Handle the request
	}
	// Request info
	requestInfo := &common.RequestInfo{Request: request, Ctx: ctx, Client: &r.client, Object: instance, Log: log}

	// Ignore ingresses not managed by this controller
	if !class.IsValid(instance) {
		log.Info("Ignore ingress based on ingress class")
		return reconcile.Result{}, nil
	}

	// TODO: (renuka) sample record
	r.recorder.Event(instance, corev1.EventTypeNormal, "SampleRecord", "Example record to test :)")

	// TODO (renuka) do not need finalizers since we are storing state in a configmap, so delete following code.
	//// Handle deletion with finalizers
	//if _, finUpdated, err := k8s.HandleDeletion(requestInfo, finalizerName, finalizeDeletion); finUpdated || err != nil {
	//	// Deletion is also handled in the below segment, hence allows the flow to continue
	//	// If finalizer updated, end the flow as a new request will queue since ingress is updated
	//	// If error should requeue request
	//	return reconcile.Result{}, err
	//}

	ingList := &v1beta1.IngressList{}
	// Read all ingresses in all namespaces
	// TODO: (renuka) add a config to handle namespace to watch (all_namespace or specific namespace)
	// watch namespace read from env variable
	if err := r.client.List(ctx, ingList, client.InNamespace("")); err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Filter out ingresses managed by the microgateway ingress controller
	ingresses := make([]*v1beta1.Ingress, 0, len(ingList.Items))
	for i := range ingList.Items {
		if class.IsValid(&ingList.Items[i]) {
			ingresses = append(ingresses, &ingList.Items[i])
		}
	}

	ingress.SortIngressSlice(ingresses)

	// Check startup
	if successfullyHandledRequestCount == len(ingresses)-1 {
		// Build the whole delta change for all ingresses
		log.Info("Build whole configurations for first time")
		if err := r.ingHandler.UpdateWholeWorld(requestInfo, ingresses); err != nil {
			return reconcile.Result{}, err
		}
		return successfullyHandled()
	} else if successfullyHandledRequestCount < len(ingresses) {
		// Ignore these requests as it will be processed in final request.
		log.Info("Ignore request")
		return successfullyHandled()
	}
	// Make incremental changes since whole world is built.

	if err := r.ingHandler.UpdateDelta(requestInfo, ingresses); err != nil {
		return reconcile.Result{}, err
	}

	// Set Ingress instance as the owner and controller
	//if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}

	// Check if this Pod already exists

	// Pod already exists - don't requeue
	//log.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return successfullyHandled()
}

func successfullyHandled() (reconcile.Result, error) {
	successfullyHandledRequestCount++
	return reconcile.Result{}, nil
}
