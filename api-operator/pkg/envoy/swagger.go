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

package envoy

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/go-openapi/loads"
	"github.com/mitchellh/mapstructure"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apim"
	"path"
	"strings"
)

const (
	EpHttp        = "http"
	EpLoadbalance = "load_balance"
	EpFailover    = "failover"
)

// generateFieldsFromSwagger3 using swagger
func getAPIData(def *apim.APIDTODefinition, swaggerDoc *loads.Document) error {
	def.Name = swaggerDoc.Spec().Info.Title
	def.Version = swaggerDoc.Spec().Info.Version
	def.Provider = "admin"
	def.Description = swaggerDoc.Spec().Info.Description
	def.Context = fmt.Sprintf("/%s/%s", def.Name, def.Version)
	def.Tags = swaggerTags(swaggerDoc)

	// fill basepath from swagger
	if swaggerDoc.BasePath() != "" {
		def.Context = path.Clean(fmt.Sprintf("/%s/%s", swaggerDoc.BasePath(), def.Version))
	}

	// override basepath if wso2 extension provided
	if basepath, ok := swaggerXWO2BasePath(swaggerDoc); ok {
		def.Context = path.Clean(basepath)
		if !strings.Contains(basepath, "{version}") {
			def.Context = path.Clean(basepath + "/" + def.Version)
			def.IsDefaultVersion = true
		} else {
			def.Context = path.Clean(strings.ReplaceAll(basepath, "{version}", def.Version))
		}
	}

	// trim spaces if available
	def.Name = strings.ReplaceAll(def.Name, " ", "")
	def.Version = strings.ReplaceAll(def.Version, " ", "")
	def.Context = strings.ReplaceAll(def.Context, " ", "")

	cors, ok, err := swagger2XWSO2Cors(swaggerDoc)
	if err != nil && ok {
		return err
	}
	if ok {
		def.CorsConfiguration = cors
	}

	prodEp, foundProdEp, err := swaggerXWSO2ProductionEndpoints(swaggerDoc)
	if err != nil && foundProdEp {
		return err
	}
	sandboxEp, foundSandboxEp, err := swaggerXWSO2SandboxEndpoints(swaggerDoc)
	if err != nil && foundSandboxEp {
		return err
	}

	if foundProdEp || foundSandboxEp {
		ep, err := BuildAPIMEndpoints(prodEp, sandboxEp)
		if err != nil {
			return err
		}
		var endpointConfig map[string]interface{}
		err = json.Unmarshal([]byte(ep), &endpointConfig)
		if err != nil {
			return err
		}
		def.EndpointConfig = &endpointConfig
	}
	return nil
}

type Tag struct {
	Name string `json:"name"`
}

func swagger2XWSO2Cors(document *loads.Document) (*apim.CorsConfiguration, bool, error) {
	if v, ok := document.Spec().Extensions["x-wso2-cors"]; ok {
		var cors apim.CorsConfiguration
		err := mapstructure.Decode(v, &cors)
		if err != nil {
			return nil, true, err
		}
		cors.CorsConfigurationEnabled = true
		return &cors, true, nil
	}
	return nil, false, nil
}

func swaggerTags(document *loads.Document) []string {
	tags := make([]string, len(document.Spec().Tags))
	for i, v := range document.Spec().Tags {
		tags[i] = v.Name
	}
	return tags
}

func swaggerXWO2BasePath(document *loads.Document) (string, bool) {
	if v, ok := document.Spec().Extensions["x-wso2-basePath"]; ok {
		str, ok := v.(string)
		return str, ok
	}
	return "", false
}

type Endpoints struct {
	Type string   `yaml:"type"`
	Urls []string `yaml:"urls"`
}

// Configuration represents endpoint config
type Configuration struct {
	// RetryTimeOut for endpoint
	RetryTimeOut *int `yaml:"retryTimeOut,omitempty" json:"retryTimeOut,omitempty"`
	// RetryDelay for endpoint
	RetryDelay *int `yaml:"retryDelay,omitempty" json:"retryDelay,omitempty"`
	// Factor used for config
	Factor *int `yaml:"factor,omitempty" json:"factor,omitempty"`
}

// Endpoint details
type Endpoint struct {
	// Type of the endpoints
	EndpointType string `json:"endpoint_type,omitempty"`
	// Url of the endpoint
	Url *string `yaml:"url" json:"url"`
	// Config of endpoint
	Config *Configuration `yaml:"config,omitempty" json:"config,omitempty"`
}

func swaggerXWSO2ProductionEndpoints(document *loads.Document) (*Endpoints, bool, error) {
	if v, ok := document.Spec().Extensions["x-wso2-production-endpoints"]; ok {
		var prodEp Endpoints
		err := mapstructure.Decode(v, &prodEp)
		if err != nil {
			return nil, true, err
		}
		return &prodEp, true, nil
	}
	return &Endpoints{}, false, nil
}

func swaggerXWSO2SandboxEndpoints(document *loads.Document) (*Endpoints, bool, error) {
	if v, ok := document.Spec().Extensions["x-wso2-sandbox-endpoints"]; ok {
		var sandboxEp Endpoints
		err := mapstructure.Decode(v, &sandboxEp)
		if err != nil {
			return nil, true, err
		}
		return &sandboxEp, true, nil
	}
	return &Endpoints{}, false, nil
}

// BuildAPIMEndpoints builds endpointConfig for given config
func BuildAPIMEndpoints(production, sandbox *Endpoints) (string, error) {
	epType := EpHttp
	if len(production.Urls) > 1 {
		epType = EpLoadbalance
		if production.Type == EpFailover {
			epType = EpFailover
		}
	}

	if len(production.Urls) == 0 {
		if len(sandbox.Urls) > 1 {
			epType = EpLoadbalance
		}
		if sandbox.Type == EpFailover {
			epType = EpFailover
		}
	}

	switch epType {
	case EpHttp:
		endpoint := buildHttpEndpoint(production, sandbox)
		return endpoint, nil
	case EpLoadbalance:
		endpoint := buildLoadBalancedEndpoints(production, sandbox)
		return endpoint, nil
	case EpFailover:
		endpoint := buildFailOver(production, sandbox)
		return endpoint, nil
	default:
		return "", fmt.Errorf("unknown endpoint type")
	}
}

func buildFailOver(production *Endpoints, sandbox *Endpoints) string {
	jsonObj, _ := gabs.ParseJSON([]byte(`
					{
						"endpoint_type": "failover",
		    			"algoCombo": "org.apache.synapse.endpoints.algorithms.RoundRobin",
		    			"algoClassName": "",
						"sessionManagement": "",
		    			"sessionTimeOut": "",
		    			"failOver": "True"
					}
				`))
	if len(production.Urls) > 0 {
		buildFailOverUrls(jsonObj, production, "production")
	}
	if len(sandbox.Urls) > 0 {
		buildFailOverUrls(jsonObj, sandbox, "sandbox")
	}
	return jsonObj.String()
}

func buildFailOverUrls(jsonObj *gabs.Container, endpoints *Endpoints, eptype string) {
	_, _ = jsonObj.Set(Endpoint{Url: &endpoints.Urls[0]}, fmt.Sprintf("%s_endpoints", eptype))
	rest := endpoints.Urls[1:]
	if len(rest) > 0 {
		fo := make([]Endpoint, len(rest))
		for i := 0; i < len(fo); i++ {
			fo[i] = Endpoint{Url: &rest[i]}
		}
		if len(fo) > 0 {
			_, _ = jsonObj.Set(fo, fmt.Sprintf("%s_failovers", eptype))
		}
	}
}

func buildLoadBalancedEndpoints(production *Endpoints, sandbox *Endpoints) string {
	jsonObj, _ := gabs.ParseJSON([]byte(`
		{
			"endpoint_type": "load_balance",
		    "algoCombo": "org.apache.synapse.endpoints.algorithms.RoundRobin",
		    "algoClassName": "org.apache.synapse.endpoints.algorithms.RoundRobin",
		    "sessionManagement": "",
		    "sessionTimeOut": ""
		}
	`))
	prodEps := make([]Endpoint, len(production.Urls))
	for i := 0; i < len(prodEps); i++ {
		prodEps[i] = Endpoint{Url: &production.Urls[i]}
	}
	if len(prodEps) > 0 {
		_, _ = jsonObj.Set(prodEps, "production_endpoints")
	}

	sandboxEps := make([]Endpoint, len(sandbox.Urls))
	for i := 0; i < len(sandboxEps); i++ {
		sandboxEps[i] = Endpoint{Url: &sandbox.Urls[i]}
	}
	if len(sandboxEps) > 0 {
		_, _ = jsonObj.Set(sandboxEps, "sandbox_endpoints")
	}

	return jsonObj.String()
}

func buildHttpEndpoint(production *Endpoints, sandbox *Endpoints) string {
	jsonObj := gabs.New()
	_, _ = jsonObj.Set(EpHttp, "endpoint_type")
	if len(production.Urls) > 0 {
		var ep Endpoint
		ep.Url = &production.Urls[0]
		_, _ = jsonObj.SetP(ep, "production_endpoints")
	}
	if len(sandbox.Urls) > 0 {
		var ep Endpoint
		ep.Url = &sandbox.Urls[0]
		_, _ = jsonObj.SetP(ep, "sandbox_endpoints")
	}
	return jsonObj.String()
}
