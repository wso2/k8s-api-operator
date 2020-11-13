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

package ratelimiting

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"

	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
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

	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
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

	userNameSpace := instance.Namespace
	//gets the details of the operator as the owner
	operatorOwner, ownerErr := getOperatorOwner(r)
	if ownerErr != nil {
		reqLogger.Info("Operator was not found in the operator namespace. No owner will be set for the artifacts",
			"operator_namespace", config.OperatorNamespace)
	}

	// GENERATE POLICY YAML USING CRD INSTANCE

	nameArray := strings.Split(instance.ObjectMeta.Name, "-")
	name := nameArray[0]
	log.Info(name)

	policyType := instance.Spec.Type
	if policyType == "subscription" || policyType == subscriptionConst {
		policyType = subscriptionConst
	} else if policyType == "application" || policyType == applicationConst {
		policyType = applicationConst
	} else if policyType == "advance" || policyType == "Advance" {
		policyType = resourceConst
	} else {
		log.Info("INVALID policy type. Use application or subscription in crd object for type")
		return reconcile.Result{}, nil
	}

	//Check if policy configmap is available
	foundmapc := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: policyConfMapNameConst, Namespace: userNameSpace}, foundmapc)

	if err != nil && errors.IsNotFound(err) {
		//create new map with default policies if a map is not found
		reqLogger.Info("Creating a config map with default policies", "Namespace", userNameSpace, "Name", policyConfMapNameConst)

		defaultval := CreateDefault()
		fmt.Println(defaultval)

		confmap, confer := CreatePolicyConfigMap(defaultval, operatorOwner, userNameSpace)
		if confer != nil {
			log.Error(confer, "Error in default config map structure creation")
		}
		foundmapc = confmap
		err = r.client.Create(context.TODO(), confmap)
		if err != nil {
			log.Error(err, "error ")
			return reconcile.Result{}, err
		}

	} else if err != nil {
		log.Error(err, "error ")
		return reconcile.Result{}, err
	}

	oldmap := foundmapc.Data
	olddata := oldmap[policyFileConst]
	count := instance.Spec.RequestCount.Limit
	unitTime := instance.Spec.UnitTime
	timeUnit := instance.Spec.TimeUnit
	policyobj := Policy{Count: count, UnitTime: unitTime, TimeUnit: timeUnit}

	oldStruct := PolicyYaml{}
	unmarshalEr := yaml.Unmarshal([]byte(olddata), &oldStruct)

	if unmarshalEr != nil {
		log.Error(unmarshalEr, "Conf map data unmarshal error")
	}

	oldRes := oldStruct.ResourcePolicies
	oldSub := oldStruct.SubscriptionPolicies
	oldApp := oldStruct.ApplicationPolicies
	var newRes []map[string]Policy
	var newSub []map[string]Policy
	var newApp []map[string]Policy

	if policyType == resourceConst {
		newRes = *(getUpdatedPolicy(oldRes, name, policyobj))
		newSub = oldSub
		newApp = oldApp
	} else if policyType == subscriptionConst {
		newRes = oldRes
		newSub = *(getUpdatedPolicy(oldSub, name, policyobj))
		newApp = oldApp
	} else if policyType == applicationConst {
		newRes = oldRes
		newSub = oldSub
		newApp = *(getUpdatedPolicy(oldApp, name, policyobj))
	}

	outbyte, yamler := yaml.Marshal(&PolicyYaml{ResourcePolicies: newRes, SubscriptionPolicies: newSub, ApplicationPolicies: newApp})
	var output string
	if yamler == nil {
		output = string(outbyte)
	} else {
		log.Error(yamler, "yaml marshal error")
	}

	fmt.Println(output)

	//CREATE CONFIG MAP OF POLICY YAML

	confmap, confEr := CreatePolicyConfigMap(output, operatorOwner, userNameSpace)
	if confEr != nil {
		log.Error(confEr, "Error in config map structure creation")
	}

	//Updating the policy yaml configmap
	reqLogger.Info("Updating Config map", "confmap.Namespace", confmap.Namespace, "confmap.Name", confmap.Name)
	err = r.client.Update(context.TODO(), confmap)
	if err != nil {
		log.Error(err, "error ")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil

}

//CreatePolicyConfigMap creates a config file with the generated code
func CreatePolicyConfigMap(output string, operatorOwner []metav1.OwnerReference, userNameSpace string) (*corev1.ConfigMap, error) {

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            policyConfMapNameConst,
			Namespace:       userNameSpace,
			OwnerReferences: operatorOwner,
		},
		Data: map[string]string{
			policyFileConst: output,
		},
	}, nil
}

// getUpdatedPolicy returns the updated policy array with given new policy
func getUpdatedPolicy(policyArrayMap []map[string]Policy, name string, newPolicy Policy) *[]map[string]Policy {
	updatePolicyArr := policyArrayMap
	newPolObj := map[string]Policy{
		name: newPolicy,
	}

	for i, policy := range policyArrayMap {
		if _, exist := policy[name]; exist {
			// apply change if it is exists
			updatePolicyArr[i] = newPolObj
			return &updatePolicyArr
		}
	}

	// add new if it is not exists
	updatePolicyArr = append(updatePolicyArr, newPolObj)
	return &updatePolicyArr
}

// CreateDefault creates the structure of policy yaml with default policies
func CreateDefault() string {
	res1 := Policy{Count: 50000, UnitTime: 1, TimeUnit: "min"}
	res2 := Policy{Count: 20000, UnitTime: 1, TimeUnit: "min"}
	res3 := Policy{Count: 10000, UnitTime: 1, TimeUnit: "min"}

	app1 := Policy{Count: 50, UnitTime: 1, TimeUnit: "min"}
	app2 := Policy{Count: 20, UnitTime: 1, TimeUnit: "min"}
	app3 := Policy{Count: 10, UnitTime: 1, TimeUnit: "min"}

	sub1 := Policy{Count: 5000, UnitTime: 1, TimeUnit: "min"}
	sub2 := Policy{Count: 2000, UnitTime: 1, TimeUnit: "min"}
	sub3 := Policy{Count: 1000, UnitTime: 1, TimeUnit: "min"}
	sub4 := Policy{Count: 500, UnitTime: 1, TimeUnit: "min"}

	res := make([]map[string]Policy, 3)
	app := make([]map[string]Policy, 3)
	sub := make([]map[string]Policy, 5)

	res[0] = map[string]Policy{
		"50kPerMin": res1,
	}

	res[1] = map[string]Policy{
		"20kPerMin": res2,
	}

	res[2] = map[string]Policy{
		"10kPerMin": res3,
	}

	app[0] = map[string]Policy{
		"50PerMin": app1,
	}

	app[1] = map[string]Policy{
		"20PerMin": app2,
	}

	app[2] = map[string]Policy{
		"10PerMin": app3,
	}

	sub[0] = map[string]Policy{
		"Gold": sub1,
	}

	sub[1] = map[string]Policy{
		"Silver": sub2,
	}

	sub[2] = map[string]Policy{
		"Bronze": sub3,
	}

	sub[3] = map[string]Policy{
		"Unauthenticated": sub4,
	}

	sub[4] = map[string]Policy{
		"Default": sub4,
	}

	polyout, yamler := yaml.Marshal(&PolicyYaml{ResourcePolicies: res, ApplicationPolicies: app, SubscriptionPolicies: sub})

	if yamler != nil {
		fmt.Println("error in creating default values ")
	}

	return string(polyout)

}

//PolicyYaml is the struct of Policy yaml
type PolicyYaml struct {
	ResourcePolicies     []map[string]Policy `yaml:"resourcePolicies"`
	ApplicationPolicies  []map[string]Policy `yaml:"applicationPolicies"`
	SubscriptionPolicies []map[string]Policy `yaml:"subscriptionPolicies"`
}

//Policy is the struct of one policy
type Policy struct {
	Count    int    `yaml:"count"`
	UnitTime int    `yaml:"unitTime"`
	TimeUnit string `yaml:"timeUnit"`
}

//gets the details of the operator for owner reference
func getOperatorOwner(r *ReconcileRateLimiting) ([]metav1.OwnerReference, error) {
	depFound := &appsv1.Deployment{}
	setOwner := true
	deperr := r.client.Get(context.TODO(), types.NamespacedName{Name: "api-operator", Namespace: config.OperatorNamespace}, depFound)
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
