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
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ingressForIntegration returns a ingress object
func (r *ReconcileIntegration) ingressForIntegration(config *EIConfigNew) *networking.Ingress {
	var m = config.integration
	var ingressConfig = config.ingressConfigMap
	var tlsSecretName = ingressConfig.Data[tlsSecretNameKey]
	var ingressHostName = ingressConfig.Data[ingressHostNameKey]
	ingressPaths := GenerateIngressPaths(&m)

	var ingressSpec networking.IngressSpec
	if tlsSecretName != "" {
		ingressSpec = networking.IngressSpec{
			TLS: []networking.IngressTLS{
				{
					Hosts:      []string{ingressHostName},
					SecretName: tlsSecretName,
				},
			},
			Rules: []networking.IngressRule{
				{
					Host: ingressHostName,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: ingressPaths,
						},
					},
				},
			},
		}
	} else {
		ingressSpec = networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: ingressHostName,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
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
	ingress := &networking.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
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

func (r *ReconcileIntegration) updateIngressForIntegration(config *EIConfigNew, currentIngress *networking.Ingress) *networking.Ingress {
	currentRules, _ := CheckIngressRulesExist(config, currentIngress)

	var ingressSpec networking.IngressSpec
	var tlsSecretName = config.ingressConfigMap.Data[tlsSecretNameKey]
	var ingressHostName = config.ingressConfigMap.Data[ingressHostNameKey]
	if tlsSecretName != "" {
		ingressSpec = networking.IngressSpec{
			TLS: []networking.IngressTLS{
				{
					Hosts:      []string{ingressHostName},
					SecretName: tlsSecretName,
				},
			},
			Rules: currentRules,
		}
	} else {
		ingressSpec = networking.IngressSpec{
			Rules: currentRules,
		}
	}
	currentIngress.Spec = ingressSpec
	return currentIngress
}
