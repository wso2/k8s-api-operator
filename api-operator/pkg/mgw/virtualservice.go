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
	"errors"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	istioapi "istio.io/api/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/yaml"
	"strings"
)

const (
	// istio config map
	istioConfMapName       = "istio-configs"
	istioGatewayConfKey    = "gatewayName"
	istioHostConfKey       = "host"
	istioCorsPolicyConfKey = "corsPolicy"
)

type IstioConfigs struct {
	GatewayName string
	Host        string
	CorsPolicy  *istioapi.CorsPolicy
}

var istioConfigs IstioConfigs
var logVsc = log.Log.WithName("mgw.virtualservice")

func IstioVirtualService(api *wso2v1alpha1.API, apiBasePathMap map[string]string, owner []metav1.OwnerReference) *istioclient.VirtualService {
	// labels
	labels := map[string]string{
		"app": api.Name,
	}

	// http route matches
	var httpRouteMatches []*istioapi.HTTPMatchRequest
	for basePath, version := range apiBasePathMap {
		// if the base path contains /petstore/{version}, then it is converted to /petstore/1.0.0
		if strings.Contains(basePath, versionField) {
			basePath = strings.Replace(basePath, versionField, version, 1)
		}

		match := &istioapi.HTTPMatchRequest{
			Uri: &istioapi.StringMatch{
				MatchType: &istioapi.StringMatch_Prefix{Prefix: basePath},
			},
		}
		httpRouteMatches = append(httpRouteMatches, match)
	}

	// HTTP routes
	httpRoutes := []*istioapi.HTTPRoute{{
		Route: []*istioapi.HTTPRouteDestination{{
			Destination: &istioapi.Destination{
				Host: api.Name, // MGW service name
				Port: &istioapi.PortSelector{
					Number: 9090,
				},
			},
		}},
		Match:      httpRouteMatches,
		CorsPolicy: istioConfigs.CorsPolicy,
	}}

	// Istio virtual service
	virtualService := istioclient.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:            api.Name,
			Namespace:       api.Namespace,
			Generation:      0,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: istioapi.VirtualService{
			Hosts:    []string{istioConfigs.Host},
			Gateways: []string{istioConfigs.GatewayName},
			Http:     httpRoutes,
		},
	}

	return &virtualService
}

// ValidateIstioConfigs validate the Istio yaml config read from config map "istio-configs"
// and setting values
func ValidateIstioConfigs(client *client.Client) error {
	istioConfigMap := k8s.NewConfMap()
	if err := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: istioConfMapName},
		istioConfigMap); err != nil {
		return err
	}

	// gateway
	if istioConfigMap.Data[istioGatewayConfKey] == "" {
		err := errors.New("istio gateway config is empty")
		logVsc.Error(err, "Istio gateway config is empty", "configmap", istioConfMapName,
			"key", istioGatewayConfKey)
		return err
	}
	istioConfigs.GatewayName = istioConfigMap.Data[istioGatewayConfKey]

	// host
	if istioConfigMap.Data[istioHostConfKey] == "" {
		err := errors.New("istio gateway host config is empty")
		logVsc.Error(err, "Istio gateway host config is empty", "configmap", istioConfMapName,
			"key", istioHostConfKey)
		return err
	}
	istioConfigs.Host = istioConfigMap.Data[istioHostConfKey]

	// CORS policy
	cors := &istioapi.CorsPolicy{}
	if err := yaml.Unmarshal([]byte(istioConfigMap.Data[istioCorsPolicyConfKey]), cors); err != nil {
		logVsc.Error(err, "Istio CORS policy configs are invalid", "configmap", istioConfigMap)
		return err
	}
	istioConfigs.CorsPolicy = cors

	return nil
}
