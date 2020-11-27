/*
 * Copyright (c) 2020 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ingressForIntegration returns a ingress object
func (r *ReconcileIntegration) ingressForIntegration(m *wso2v1alpha1.Integration, eic *EIController) *v1beta1.Ingress {
	ingressPaths := GenerateIngressPaths(m)

	var ingressSpec v1beta1.IngressSpec
	if eic.TLS != "" {
		ingressSpec = v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts:      []string{eic.Host},
					SecretName: eic.TLS,
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: eic.Host,
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
					Host: eic.Host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: ingressPaths,
						},
					},
				},
			},
		}
	}

	// create ingress object using provided or default values
	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameForIngress(),
			Namespace: m.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                "nginx",
				"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
				"nginx.ingress.kubernetes.io/ssl-redirect":   eic.SSLRedirect,
			},
		},
		Spec: ingressSpec,
	}
	return ingress
}

func (r *ReconcileIntegration) updateIngressForIntegration(m *wso2v1alpha1.Integration, eic *EIController, currentIngress *v1beta1.Ingress) *v1beta1.Ingress {
	currentRules, _ := CheckIngressRulesExist(m, eic, currentIngress)

	var ingressSpec v1beta1.IngressSpec
	if eic.TLS != "" {
		ingressSpec = v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts:      []string{eic.Host},
					SecretName: eic.TLS,
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
