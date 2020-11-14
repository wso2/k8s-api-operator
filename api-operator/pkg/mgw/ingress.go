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
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var loggerIng = log.Log.WithName("mgw.ingress")

const (
	ingressConfigs       = "ingress-configs"
	ingressResourceName  = "ingressResourceName"
	ingressTransportMode = "ingressTransportMode"
	ingressHostName      = "ingressHostName"
	tlsSecretName        = "tlsSecretName"
	ingressProperties    = "ingress.properties"

	versionField = "{version}"
)

// ApplyIngressResource creates or updates an Ingress resource to expose mgw
// Supports for multiple apiBasePaths when there are multiple swaggers for one API CRD
func ApplyIngressResource(client *client.Client, api *wso2v1alpha1.API, apiBasePathMap map[string]string, owner *[]metav1.OwnerReference) error {
	logIng := loggerIng.WithValues("namespace", api.Namespace, "apiName", api.Name)
	ingressConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: ingressConfigs}, ingressConfMap)
	if err != nil {
		logIng.Error(err, "Error retrieving ingress configmap", "name", ingressConfigs)
		return err
	}

	transportMode := ingressConfMap.Data[ingressTransportMode]
	var ingressHostname string
	if api.Spec.IngressHostname != "" {
		ingressHostname = api.Spec.IngressHostname
	} else {
		ingressHostname = ingressConfMap.Data[ingressHostName]
	}
	tlsSecretName := ingressConfMap.Data[tlsSecretName]
	ingressNamePrefix := ingressConfMap.Data[ingressResourceName]
	ingressName := ingressNamePrefix + "-" + api.Name
	namespace := api.Namespace
	apiServiceName := api.Name

	hostArray := []string{ingressHostname}
	logIng.Info("Creating ingress resource ", "name", ingressName,
		"transport_mode", transportMode, "ingress_host_name", ingressHostname)

	var port int32
	if httpConst == transportMode {
		port = Configs.HttpPort
	} else {
		port = Configs.HttpsPort
	}

	ingressAnnotationMap := make(map[string]string)
	splitArray := strings.Split(ingressConfMap.Data[ingressProperties], "\n")
	for _, element := range splitArray {
		if element != "" && strings.ContainsAny(element, ":") {
			splitValues := strings.Split(element, ":")
			ingressAnnotationMap[strings.TrimSpace(splitValues[0])] = strings.TrimSpace(splitValues[1])
		}
	}

	logIng.Info("Creating ingress resource with the following Base Paths")

	// add multiple api base paths
	var httpIngressPaths []v1beta1.HTTPIngressPath
	for basePath, version := range apiBasePathMap {
		// if the base path contains /petstore/{version}, then it is converted to /petstore/1.0.0
		if strings.Contains(basePath, versionField) {
			basePath = strings.Replace(basePath, versionField, version, -1)
		}

		logIng.Info("Adding the base path to ingress resource", "base_path", basePath)
		httpIngressPaths = append(httpIngressPaths, v1beta1.HTTPIngressPath{
			Path: basePath,
			Backend: v1beta1.IngressBackend{
				ServiceName: apiServiceName,
				ServicePort: intstr.IntOrString{IntVal: port},
			},
		})
	}

	ingressResource := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       namespace, // goes into backend full name
			Name:            ingressName,
			Annotations:     ingressAnnotationMap,
			OwnerReferences: *owner,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressHostname,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: httpIngressPaths,
						},
					},
				},
			},
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      hostArray,
					SecretName: tlsSecretName,
				},
			},
		},
	}

	return k8s.Apply(client, ingressResource)
}
