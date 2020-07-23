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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logEp = logf.Log.WithName("endpoint.value")

func ExternalIP (client *client.Client, apiInstance *wso2v1alpha1.API, operatorMode string, svc *corev1.Service,
	ingressConfData map[string]string, openshiftConfData map[string]string) string {

	logger := logEp.WithValues("namespace", apiInstance.Namespace, "apiName", apiInstance.Name)
	var ip string
	if operatorMode == "default" {
		loadBalancerFound := svc.Status.LoadBalancer.Ingress
		ip = ""
		for _, elem := range loadBalancerFound {
			ip += elem.IP
		}
		apiInstance.Spec.ApiEndPoint = ip
		logger.Info("IP value is :" + ip)
		logger.Info("ENDPOINT value in default mode is ","apiEndpoint",apiInstance.Spec.ApiEndPoint)
	}
	if operatorMode == "ingress" {
		ingressHostConf := ingressConfData[ingressHostName]
		ingResource := &v1beta1.Ingress{}
		errRes := k8s.Get(client,
			types.NamespacedName{Namespace: apiInstance.Namespace, Name: ingressConfData[ingressResourceName] + "-" + apiInstance.Name},
			ingResource)
		if errRes != nil {
			logger.Error(errRes, "Error getting the Ingress resources")
		} else {
			ingressIPFound := ingResource.Status.LoadBalancer.Ingress
			for _, elem := range ingressIPFound {
				ip += elem.IP
			}
		}
		logger.Info("Ingress IP is: " + ip)
		logger.Info("Host Name is: " + ingressHostConf)
		apiInstance.Spec.ApiEndPoint = ingressHostConf + ", " + ip
		logger.Info("ENDPOINT value in ingress mode is","apiEndpoint",apiInstance.Spec.ApiEndPoint)
	}
	if operatorMode == "route" {
		routeHostConf := openshiftConfData[routeHost]
		logger.Info("Host Name is :" + routeHostConf)
		apiInstance.Spec.ApiEndPoint = routeHostConf
		logger.Info("ENDPOINT value in route mode is","apiEndpoint",apiInstance.Spec.ApiEndPoint)
		ip = "<pending>"
	}
	if apiInstance.Spec.ApiEndPoint == "" {
		apiInstance.Spec.ApiEndPoint = "<pending>"
		logger.Info("ENDPOINT value after updating is","apiEndpoint" ,apiInstance.Spec.ApiEndPoint)
	}

	return ip
}
