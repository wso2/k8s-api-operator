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
	routv1 "github.com/openshift/api/route/v1"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var loggerRoute = log.Log.WithName("mgw.route")

const (
	openShiftConfigs = "route-configs"

	routeName          = "routeName"
	routeHost          = "routeHost"
	routeTransportMode = "routeTransportMode"
	tlsTermination     = "tlsTermination"

	edge            = "edge"
	reEncrypt       = "reencrypt"
	passThrough     = "passthrough"
	routeProperties = "route.properties"
	serviceKind     = "Service"
)

// ApplyRouteResource creates or updates a route resource to expose MGW
// Supports for multiple apiBasePaths when there are multiple swaggers for one API CRD
func ApplyRouteResource(client *client.Client, api *wso2v1alpha1.API,
	apiBasePathMap map[string]string, owner *[]metav1.OwnerReference) error {
	logRoute := loggerRoute.WithValues("namespace", api.Namespace, "apiName", api.Name)
	routeConfMap := k8s.NewConfMap()
	errRoute := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: openShiftConfigs}, routeConfMap)
	if errRoute != nil {
		logRoute.Error(errRoute, "Error retrieving route configmap")
		return errRoute
	}

	routePrefix := routeConfMap.Data[routeName]
	routesHostname := routeConfMap.Data[routeHost]
	transportMode := routeConfMap.Data[routeTransportMode]
	tlsTerminationValue := routeConfMap.Data[tlsTermination]

	var tlsTerminationType routv1.TLSTerminationType
	if strings.EqualFold(tlsTerminationValue, edge) {
		tlsTerminationType = routv1.TLSTerminationEdge
	} else if strings.EqualFold(tlsTerminationValue, reEncrypt) {
		tlsTerminationType = routv1.TLSTerminationReencrypt
	} else if strings.EqualFold(tlsTerminationValue, passThrough) {
		tlsTerminationType = routv1.TLSTerminationPassthrough
	} else {
		tlsTerminationType = ""
	}

	routeName := routePrefix + "-" + api.Name
	namespace := api.Namespace
	apiServiceName := api.Name

	logRoute.Info("Creating route resource", "name", routeName, "transport_mode", transportMode,
		"routes_host_name", routesHostname)

	var port int32
	if httpConst == transportMode {
		port = Configs.HttpPort
	} else {
		port = Configs.HttpsPort
	}

	annotationsList := routeConfMap.Data[routeProperties]
	routeAnnotationMap := make(map[string]string)
	splitArray := strings.Split(annotationsList, "\n")
	for _, element := range splitArray {
		if element != "" && strings.ContainsAny(element, ":") {
			splitValues := strings.Split(element, ":")
			routeAnnotationMap[strings.TrimSpace(splitValues[0])] = strings.TrimSpace(splitValues[1])
		}
	}

	logRoute.Info("Creating route resource for API", "api_name", api.Name)
	var routeList []routv1.Route

	for basePath := range apiBasePathMap {
		apiBasePath := basePath
		// if the base path contains /store/{version}, then it is converted to /store/1.0.0
		if strings.Contains(basePath, versionField) {
			apiBasePath = strings.Replace(basePath, versionField, apiBasePathMap[basePath], -1)
		}

		apiBasePathSuffix := apiBasePath
		apiBasePathSuffix = strings.Replace(apiBasePathSuffix, "/", "-", -1)
		routeNewName := routeName + apiBasePathSuffix

		logRoute.Info("Creating the route to ingress resource", "api_base_path", apiBasePath)
		routeResource := routv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:            routeNewName,
				Namespace:       namespace,
				OwnerReferences: *owner,
				Annotations:     routeAnnotationMap,
			},
			Spec: routv1.RouteSpec{
				Host: routesHostname,
				Path: apiBasePath,
				Port: &routv1.RoutePort{
					TargetPort: intstr.IntOrString{IntVal: port},
				},
				To: routv1.RouteTargetReference{
					Kind: serviceKind,
					Name: apiServiceName,
				},
				TLS: &routv1.TLSConfig{
					Termination: tlsTerminationType,
				},
			},
		}

		routeList = append(routeList, routeResource)
	}

	for _, route := range routeList {
		if err := k8s.Apply(client, &route); err != nil {
			return err
		}
	}

	return nil
}
