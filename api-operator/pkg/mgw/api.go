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
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logEp = logf.Log.WithName("endpoint.value")

func ExternalIP(client *client.Client, apiInstance *wso2v1alpha1.API, operatorMode string, svc *corev1.Service,
	ingressConfData map[string]string, openshiftConfData map[string]string, istioConfigs *IstioConfigs) string {

	logger := logEp.WithValues("namespace", apiInstance.Namespace, "apiName", apiInstance.Name)
	ipList := make(map[string]bool, 2) // to avoid duplicate IPs make ipList a map of strings -> bool

	// default mode
	if strings.EqualFold(operatorMode, defaultMode) {
		ingresses := svc.Status.LoadBalancer.Ingress
		for _, ingress := range ingresses {
			if ingress.IP != "" {
				ipList[ingress.IP] = true
			} else if ingress.Hostname != "" {
				ipList[ingress.Hostname] = true
			}
		}
	}

	// ingress mode
	if strings.EqualFold(operatorMode, ingressMode) {
		ingResource := &v1beta1.Ingress{}
		errRes := k8s.Get(client,
			types.NamespacedName{Namespace: apiInstance.Namespace, Name: ingressConfData[ingressResourceName] + "-" + apiInstance.Name},
			ingResource)
		if errRes != nil {
			logger.Error(errRes, "Error getting the Ingress resources")
		} else {
			ingressRules := ingResource.Spec.Rules
			for _, rule := range ingressRules {
				if rule.Host != "" {
					ipList[rule.Host] = true
				}
			}

			ingresses := ingResource.Status.LoadBalancer.Ingress
			for _, ingress := range ingresses {
				if ingress.IP != "" {
					ipList[ingress.IP] = true
				} else if ingress.Hostname != "" {
					ipList[ingress.Hostname] = true
				}
			}
		}
	}

	// route mode
	if strings.EqualFold(operatorMode, routeMode) {
		routeHostConf := openshiftConfData[routeHost]
		ipList[routeHostConf] = true
	}

	// istio mode
	if strings.EqualFold(operatorMode, IstioMode) {
		ipList[istioConfigs.Host] = true
	}

	ips := make([]string, 0, len(ipList))
	for ip := range ipList {
		ips = append(ips, ip)
	}

	// set ip to api instance
	ipString := strings.Join(ips, ", ")
	apiInstance.Spec.ApiEndPoint = ipString

	// ip not found
	if ipString == "" {
		apiInstance.Spec.ApiEndPoint = "<pending>"
	}

	logger.Info("Setting API endpoint value", "api.endpoint", ipString)
	return ipString
}

// get hostAliases for the deployment
func getHostAliases(client *client.Client) []corev1.HostAlias {
	mgwDeploymentConfMap := k8s.NewConfMap()
	errGetDeploy := k8s.Get(client, types.NamespacedName{Name: mgwDeploymentConfigMapName, Namespace: config.SystemNamespace},
		mgwDeploymentConfMap)
	if errGetDeploy != nil {
		logEp.Error(errGetDeploy, "Error getting mgw deployment configs")
	}

	var aliases []corev1.HostAlias
	yamlErrDeploymentConfigMaps := yaml.Unmarshal([]byte(mgwDeploymentConfMap.Data["hostAliases"]), &aliases)
	if yamlErrDeploymentConfigMaps != nil {
		logEp.Error(yamlErrDeploymentConfigMaps, "Error setting the hostAliases")
	}

	return aliases
}
