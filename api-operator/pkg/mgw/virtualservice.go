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
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
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
	istiotlsConfKey        = "tls"
	istioCorsPolicyConfKey = "corsPolicy"
)

type IstioConfigs struct {
	GatewayName string
	Host        string
	CorsPolicy  *istioapi.CorsPolicy
	Tls         *tlsRoutesConfigs
}

type tlsRoutesConfigs struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

var logVsc = log.Log.WithName("mgw.virtualservice")

func IstioVirtualService(istioConfigs *IstioConfigs, api *wso2v1alpha1.API, apiBasePathMap map[string]string,
	owner []metav1.OwnerReference) *istioclient.VirtualService {
	// labels
	labels := map[string]string{
		"app": api.Name,
	}

	// route mode TLS/HTTP
	var httpRoutes []*istioapi.HTTPRoute
	var tlsRoutes []*istioapi.TLSRoute

	// select route mode TLS/HTTP
	if istioConfigs.Tls.Enabled { // TLS mode
		tlsRoutes = getTlsRoutes(istioConfigs, api)
	} else { // HTTP mode
		httpRoutes = getHttpRoutes(istioConfigs, api, apiBasePathMap)
	}

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
			Tls:      tlsRoutes,
		},
	}

	return &virtualService
}

func getHttpRoutes(istioConfigs *IstioConfigs, api *wso2v1alpha1.API, apiBasePathMap map[string]string) []*istioapi.HTTPRoute {
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
					Number: uint32(Configs.HttpPort),
				},
			},
		}},
		Match:      httpRouteMatches,
		CorsPolicy: istioConfigs.CorsPolicy,
	}}

	return httpRoutes
}

func getTlsRoutes(istioConfigs *IstioConfigs, api *wso2v1alpha1.API) []*istioapi.TLSRoute {
	tlsRoutes := []*istioapi.TLSRoute{
		{
			Match: []*istioapi.TLSMatchAttributes{{
				SniHosts: []string{istioConfigs.Host},
				Port:     0,
			}},
			Route: []*istioapi.RouteDestination{{
				Destination: &istioapi.Destination{
					Host: api.Name, // MGW service name
					Port: &istioapi.PortSelector{
						Number: uint32(Configs.HttpsPort),
					},
				},
			}},
		},
	}

	return tlsRoutes
}

// ValidateIstioConfigs validate the Istio yaml config read from config map "istio-configs"
// and setting values
func ValidateIstioConfigs(client *client.Client, api *wso2v1alpha1.API) (*IstioConfigs, error) {
	istioConfigs := &IstioConfigs{}

	istioConfigMap := k8s.NewConfMap()
	if err := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: istioConfMapName},
		istioConfigMap); err != nil {
		logVsc.Error(err, "Istio configs configmap is empty", "configmap", istioConfMapName,
			"key", istioGatewayConfKey)
		return nil, err
	}

	// gateway
	if istioConfigMap.Data[istioGatewayConfKey] == "" {
		err := errors.New("istio gateway config is empty")
		logVsc.Error(err, "Istio gateway config is empty", "configmap", istioConfMapName,
			"key", istioGatewayConfKey)
		return nil, err
	}
	istioConfigs.GatewayName = istioConfigMap.Data[istioGatewayConfKey]

	// host
	// set host from API spec if given or from configmap
	if api.Spec.IngressHostname != "" {
		istioConfigs.Host = api.Spec.IngressHostname
	} else if istioConfigMap.Data[istioHostConfKey] == "" {
		err := errors.New("istio gateway host config is empty")
		logVsc.Error(err, "Istio gateway host config is empty", "configmap", istioConfigMap,
			"key", istioHostConfKey)
		return nil, err
	} else {
		istioConfigs.Host = istioConfigMap.Data[istioHostConfKey]
	}

	// TLS
	if istioConfigMap.Data[istiotlsConfKey] == "" {
		err := errors.New("istio tls config is empty")
		logVsc.Error(err, "Istio tls config is empty", "configmap", istioConfigMap,
			"key", istioGatewayConfKey)
		return nil, err
	}
	tlsConf := &tlsRoutesConfigs{}
	if err := yaml.Unmarshal([]byte(istioConfigMap.Data[istiotlsConfKey]), tlsConf); err != nil {
		logVsc.Error(err, "Istio tls config are invalid", "configmap", istioConfigMap,
			"key", istiotlsConfKey)
		return nil, err
	}
	istioConfigs.Tls = tlsConf

	// CORS policy
	cors := &istioapi.CorsPolicy{}
	if err := yaml.Unmarshal([]byte(istioConfigMap.Data[istioCorsPolicyConfKey]), cors); err != nil {
		logVsc.Error(err, "Istio CORS policy configs are invalid", "configmap", istioConfigMap,
			"key", istioCorsPolicyConfKey)
		return nil, err
	}
	istioConfigs.CorsPolicy = cors

	return istioConfigs, nil
}
