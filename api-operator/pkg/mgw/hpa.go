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

package mgw

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/yaml"
	"strconv"
)

var logHpa = log.Log.WithName("mgw.hpa")
var metricsHpa *[]v2beta2.MetricSpec
var hpaMaxReplicas int32

const (
	hpaConfigMapName     = "hpa-configs"
	metricsConfigKey     = "metrics"
	maxReplicasConfigKey = "maxReplicas"
)

// HPA returns a HPA instance with specified config values
func HPA(api *wso2v1alpha1.API, dep *appsv1.Deployment, owner *[]metav1.OwnerReference) *v2beta2.HorizontalPodAutoscaler {
	// target resource
	targetResource := v2beta2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       dep.Name,
		APIVersion: "apps/v1",
	}

	// min replicas
	minReplicas := int32(api.Spec.Replicas)

	// HPA instance
	return &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dep.Name,
			Namespace:       dep.Namespace,
			OwnerReferences: *owner,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MinReplicas:    &minReplicas,
			MaxReplicas:    hpaMaxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        *metricsHpa,
		},
	}
}

// ValidateHpaConfigs validate the HPA yaml config read from config map "hpa-configs"
// and setting values
func ValidateHpaConfigs(client *client.Client) error {
	// get global hpa configs, return error if not found (required config map)
	hpaConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: hpaConfigMapName}, hpaConfMap)
	if err != nil {
		return err
	}

	// set max replica count
	maxReplicasInt64, errInt := strconv.ParseInt(hpaConfMap.Data[maxReplicasConfigKey], 10, 32)
	if errInt != nil {
		logHpa.Error(err, "Error parsing HPA MaxReplicas",
			"value", hpaConfMap.Data[maxReplicasConfigKey])
		return err
	}
	hpaMaxReplicas = int32(maxReplicasInt64)

	// parse hpa config yaml
	metricsHpa = &[]v2beta2.MetricSpec{}
	yamlErr := yaml.Unmarshal([]byte(hpaConfMap.Data[metricsConfigKey]), metricsHpa)
	if yamlErr != nil {
		logHpa.Error(err, "Error marshalling HPA config yaml", "configmap", hpaConfMap)
		return yamlErr
	}

	return nil
}
