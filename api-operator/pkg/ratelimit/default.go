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
