package ratelimit

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/ratelimiting"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("ratelimit")

const (
	policyConfigmap    = "policy-configmap"
	policyFileConst    = "policies.yaml"
	policyYamlLocation = "/usr/wso2/policy/"
)

// Handle handles rate limit with adding volumes to the Kaniko job
func Handle(client *client.Client, userNameSpace string, operatorOwner *[]metav1.OwnerReference) error {
	//Check if policy configmap is available
	policyConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Name: policyConfigmap, Namespace: userNameSpace}, policyConfMap)

	// if there aren't any ratelimiting objects deployed, new policy.yaml configmap will be created with default policies
	if err != nil && errors.IsNotFound(err) {
		//create new map with default policies in user namespace if a map is not found
		logger.Info("Creating policy configmap with default policies", "namespace", userNameSpace, "name", policyConfigmap)

		defaultPolicy := ratelimiting.CreateDefault() // TODO: rnk this method should come to this package. calling a controller package is not good
		policyDataMap := map[string]string{policyFileConst: defaultPolicy}
		policyConfMap := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNameSpace, Name: policyConfigmap}, &policyDataMap, nil, operatorOwner)

		if err = k8s.Create(client, policyConfMap); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	kaniko.AddVolume(k8s.ConfigMapVolumeMount(policyConfigmap, policyYamlLocation))
	return nil
}
