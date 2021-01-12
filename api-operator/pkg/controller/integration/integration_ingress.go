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
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ingressForIntegration returns a ingress object
func (r *ReconcileIntegration) ingressForIntegration(config *EIConfigNew) *v1beta1.Ingress {
	var m = config.integration
	var ingressConfig = config.ingressConfigMap
	var tlsSecretName = ingressConfig.Data[tlsSecretNameKey]
	var ingressHostName = ingressConfig.Data[ingressHostNameKey]
	ingressPaths := GenerateIngressPaths(&m)

	var ingressSpec v1beta1.IngressSpec
	if tlsSecretName != "" {
		ingressSpec = v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{ingressHostName},
					SecretName: tlsSecretName,
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressHostName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: ingressPaths,
						},
					},
				},
			},
		}
	} else {
		ingressSpec = v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressHostName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: ingressPaths,
						},
					},
				},
			},
		}
	}

	//Read ingress annotations
	ingressAnnotationMap := make(map[string]string)
	splitArray := strings.Split(ingressConfig.Data[ingressProperties], "\n")
	for _, element := range splitArray {
		if element != "" && strings.ContainsAny(element, ":") {
			splitValues := strings.Split(element, ":")
			ingressAnnotationMap[strings.TrimSpace(splitValues[0])] = strings.TrimSpace(splitValues[1])
		}
	}

	// create ingress object using provided or default values
	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        nameForIngress(),
			Namespace:   m.Namespace,
			Annotations: ingressAnnotationMap,
		},
		Spec: ingressSpec,
	}
	return ingress
}

func (r *ReconcileIntegration) updateIngressForIntegration(config *EIConfigNew, currentIngress *v1beta1.Ingress) *v1beta1.Ingress {
	currentRules, _ := CheckIngressRulesExist(config, currentIngress)

	var ingressSpec v1beta1.IngressSpec
	var tlsSecretName = config.ingressConfigMap.Data[tlsSecretNameKey]
	var ingressHostName = config.ingressConfigMap.Data[ingressHostNameKey]
	if tlsSecretName != "" {
		ingressSpec = v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{ingressHostName},
					SecretName: tlsSecretName,
				},
			},
			Rules: currentRules,
		}
	} else {
		ingressSpec = v1beta1.IngressSpec{
			Rules: currentRules,
		}
	}
	currentIngress.Spec = ingressSpec
	return currentIngress
}
