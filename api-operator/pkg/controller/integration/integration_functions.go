/*
 * Copyright (c) 2021 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package integration

import (
    "context"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"
)

// labelsForIntegration returns the labels for selecting the resources
// belonging to the given integration CR name.
func labelsForIntegration(name string) map[string]string {
	return map[string]string{"app": "integration", "integration_cr": name}
}

// nameForDeployment gives the name for the deployment
func nameForDeployment(m *wso2v1alpha2.Integration) string {
	return m.Name + deploymentNamePostfix
}

// nameForService gives the name for the service
func nameForService(m *wso2v1alpha2.Integration) string {
	return m.Name + serviceNamePostfix
}

// nameForHPA gives the name for the HPA instance
func nameForHPA(m *wso2v1alpha2.Integration) string {
	return m.Name + hpaNamePostfix
}

// nameForInboundService gives the name for the inbound service
func nameForInboundService(m *wso2v1alpha2.Integration) string {
	return m.Name + inboundServicePostfix
}

// nameForIngress gives the name for the ingress
func nameForIngress() string {
	return eiIngressName
}

// CheckIngressRulesExist checks the ingress rules are exist in current ingress
func CheckIngressRulesExist(config *EIConfigNew, currentIngress *v1beta1.Ingress) ([]v1beta1.IngressRule, bool) {
	var integration = config.integration
	ingressPaths := GenerateIngressPaths(&integration)

	currentRules := currentIngress.Spec.Rules
	newRule := v1beta1.IngressRule{
		Host: config.ingressConfigMap.Data[ingressHostNameKey],
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: ingressPaths,
			},
		},
	}

	// check the rules are exists in the ingress, if not add the rules
	// checking because of reconsile is looping
	ruleExists := false
	for _, rule := range currentRules {
		if reflect.DeepEqual(rule, newRule) {
			ruleExists = true
		}
	}

	if !ruleExists {
		currentRules = append(currentRules, newRule)
	}
	return currentRules, ruleExists
}

// GenerateIngressPaths generates the ingress paths
func GenerateIngressPaths(m *wso2v1alpha2.Integration) []v1beta1.HTTPIngressPath {
	var ingressPaths []v1beta1.HTTPIngressPath

	//Set HTTP ingress path
	httpPath := "/" + nameForService(m) + "(/|$)(.*)"
	pathType := v1beta1.PathTypeImplementationSpecific
	httpIngressPath := v1beta1.HTTPIngressPath{
		Path:     httpPath,
		PathType: &pathType,
		Backend: v1beta1.IngressBackend{
			ServiceName: nameForService(m),
			ServicePort: intstr.IntOrString{
				Type:   Int,
				IntVal: m.Spec.Expose.PassthroPort,
			},
		},
	}
	ingressPaths = append(ingressPaths, httpIngressPath)

	// check inbound endpoint port is exist and update the ingress path
	for _, port := range m.Spec.Expose.InboundPorts {
		inboundPath := "/" + nameForInboundService(m) +
			"/" + strconv.Itoa(int(port)) + "(/|$)(.*)"
		inboundIngressPath := v1beta1.HTTPIngressPath{
			Path: inboundPath,
			Backend: v1beta1.IngressBackend{
				ServiceName: nameForService(m),
				ServicePort: intstr.IntOrString{
					Type:   Int,
					IntVal: port,
				},
			},
		}
		ingressPaths = append(ingressPaths, inboundIngressPath)
	}

	return ingressPaths
}

// Get configmap by the given name
func (r *ReconcileIntegration) GetConfigMap(integration *wso2v1alpha2.Integration, configMapName string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: config.SystemNamespace}, configMap)
	if err != nil {
		log.Error(err, "Error getting the ConfigMap " + configMapName)
	}
	return configMap, err
}

//gets the details of the targetEndPoint crd object for owner reference
func getOwnerDetails(cr wso2v1alpha2.Integration) []metav1.OwnerReference {
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
