package targetendpoint

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"

	wso2v1alpha1 "github.com/wso2/k8s-apim-operator/apim-operator/pkg/apis/wso2/v1alpha1"

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

	if instance.Spec.Deploy.DockerImage != "" {
		if err := r.reconcileDeployment(instance); err != nil {
			return reconcile.Result{}, err
		}

		if err := r.reconcileService(instance); err != nil {
			return reconcile.Result{}, err
		}

	}
	return reconcile.Result{Requeue: true}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *wso2v1alpha1.TargetEndpoint) *corev1.Pod {
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

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileTargetEndpoint) newDeploymentForCR(m *wso2v1alpha1.TargetEndpoint) *appsv1.Deployment {
	replicas := m.Spec.Deploy.Count
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
					}},
				},
			},
		},
	}
	// Set Examplekind instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep

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

func (r *ReconcileTargetEndpoint) reconcileDeployment(m *wso2v1alpha1.TargetEndpoint) error {
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeploymentForCR(m)
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

// NewService assembles the ClusterIP service for the Nginx
func (r *ReconcileTargetEndpoint) newServiceForCR(m *wso2v1alpha1.TargetEndpoint) *corev1.Service {
	var port int
	port = int(m.Spec.Port)
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
				corev1.ServicePort{Port: m.Spec.Port, TargetPort: intstr.FromInt(port)},
			},
		},
	}
	controllerutil.SetControllerReference(m, &service, r.scheme)
	return &service
}
