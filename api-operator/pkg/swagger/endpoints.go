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

package swagger

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logEp = log.Log.WithName("swagger.endpoints")

const (
	ServerLess = "serverless"
	Sidecar    = "sidecar"
	privateJet = "privateJet"
)

//XMgwProductionEndpoint represents the structure of production endpoint
type XMgwProductionEndpoint struct {
	Urls []string `yaml:"urls" json:"urls"`
}

// HandleMgwEndpoints gets endpoint from swagger and replace it with targetendpoint kind service endpoint and
// returns a map of sidecar endpoints
func HandleMgwEndpoints(client *client.Client, swagger *openapi3.Swagger, mode string, apiNamespace string) (
	map[string]bool, error) {
	// map endpoint name -> exists (bool)
	sideCarEndpoints := make(map[string]bool)

	genericEndpointsSwagger, checkEndpoint := swagger.Extensions[EndpointExtension]
	if checkEndpoint {
		logEp.Info("Parsing x-wso2-endpoints extension in swagger definition")
		endpointGen, checkJsonRawNew := genericEndpointsSwagger.(json.RawMessage)
		if !checkJsonRawNew {
			err := errs.New("value is not a json.RawMessage")
			logEp.Error(err,
				"Invalid format of Target Endpoint definition in swagger")
			return nil, err
		}

		var genericEndpoints []map[string]interface{}
		if err := json.Unmarshal(endpointGen, &genericEndpoints); err != nil {
			logEp.Error(err, "Invalid format of Target Endpoint definition in swagger")
		}

		// updated endpoints
		finalisedEndpoints := genericEndpoints

		for index, _ := range genericEndpoints {
			for name, value := range genericEndpoints[index] {
				urlList := value.(map[string]interface{})["urls"].([]interface{})
				stringUrlList := make([]string, len(urlList))
				for index, epVal := range urlList {
					stringUrlList[index] = epVal.(string)
				}
				updatedEndpoint := XMgwProductionEndpoint{Urls: stringUrlList}

				errEndpoint := processEndpoints(client, updatedEndpoint.Urls, sideCarEndpoints, updatedEndpoint, apiNamespace, mode)
				if errEndpoint != nil {
					return nil, errEndpoint
				}

				finalisedEndpoints[index][name].(map[string]interface{})["urls"] = updatedEndpoint.Urls
				fmt.Println(updatedEndpoint)
			}
		}

		swagger.Extensions[EndpointExtension] = finalisedEndpoints
	}

	// API level endpoint
	if err := updateSwaggerWithProdEPs(client, swagger.Extensions, sideCarEndpoints, apiNamespace, mode); err != nil {
		return nil, err
	}

	//  Resource level endpoint
	for _, path := range swagger.Paths {
		if path.Get != nil {
			err := updateSwaggerWithProdEPs(client, path.Get.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Post != nil {
			err := updateSwaggerWithProdEPs(client, path.Post.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Put != nil {
			err := updateSwaggerWithProdEPs(client, path.Put.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Delete != nil {
			err := updateSwaggerWithProdEPs(client, path.Delete.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Patch != nil {
			err := updateSwaggerWithProdEPs(client, path.Patch.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Head != nil {
			err := updateSwaggerWithProdEPs(client, path.Head.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
		if path.Options != nil {
			err := updateSwaggerWithProdEPs(client, path.Options.Extensions, sideCarEndpoints, apiNamespace, mode)
			if err != nil {
				return nil, err
			}
		}
	}

	return sideCarEndpoints, nil
}

// updateSwaggerWithProdEPs replaces production endpoints with Target Endpoints CR values
func updateSwaggerWithProdEPs(client *client.Client, swaggerExtensions map[string]interface{},
	sideCarEndpoints map[string]bool, apiNamespace string, mode string) error {
	swaggerEpAPI, checkEndpoint := swaggerExtensions[ProductionEndpointExtension]
	// if not production endpoints defined return
	if !checkEndpoint {
		return nil
	}

	// check json format
	endpointJson, checkJsonRaw := swaggerEpAPI.(json.RawMessage)
	if !checkJsonRaw {
		err := errs.New("value is not a json.RawMessage")
		logEp.Error(err,
			"Invalid format of Target Endpoint definition in swagger")
		return err
	}

	prodEpStr := string(endpointJson)
	// check if endpoint is a reference
	if strings.EqualFold(prodEpStr[1:2], "#") {
		swaggerExtensions[ProductionEndpointExtension] = prodEpStr[1 : len(prodEpStr)-1]
		return nil
	}

	prodEp := XMgwProductionEndpoint{}
	if err := json.Unmarshal(endpointJson, &prodEp); err != nil {
		logEp.Error(err, "Invalid format of Target Endpoint definition in swagger")
	}

	// Updated URLs
	updatedEndpoint := XMgwProductionEndpoint{Urls: make([]string, len(prodEp.Urls))}

	errProdEndpoint := processEndpoints(client, prodEp.Urls, sideCarEndpoints, updatedEndpoint, apiNamespace, mode)
	if errProdEndpoint != nil {
		return errProdEndpoint
	}

	// update swagger definition
	swaggerExtensions[ProductionEndpointExtension] = updatedEndpoint
	return nil
}

// updating the endpoint values based on targetendpoints
func processEndpoints(client *client.Client, endpointList []string,
	sideCarEndpoints map[string]bool, updatedEndpoint XMgwProductionEndpoint, apiNamespace string, mode string) error {

	for index, prodEpVal := range endpointList {
		_, errUrl := url.ParseRequestURI(prodEpVal)
		if errUrl == nil { // Target EP is a valid URL
			updatedEndpoint.Urls[index] = prodEpVal
		} else { // Target EP is a name of Target EP CR
			epNamespace := apiNamespace // namespace of the endpoint
			if namespacedEp := strings.Split(prodEpVal, "."); len(namespacedEp) == 2 {
				epNamespace = namespacedEp[1]
				prodEpVal = namespacedEp[0]
			}

			targetEpCr := &wso2v1alpha1.TargetEndpoint{} // CR of the Target Endpoint
			erCr := k8s.Get(client, types.NamespacedName{Namespace: epNamespace, Name: prodEpVal}, targetEpCr)
			if erCr != nil {
				return erCr
			}

			protocol := targetEpCr.Spec.ApplicationProtocol
			port := strconv.Itoa(int(targetEpCr.Spec.Ports[0].Port))
			if strings.EqualFold(mode, Sidecar) { // sidecar mode
				sidecarUrl := fmt.Sprintf("%v://localhost:%v", protocol, port)
				sideCarEndpoints[prodEpVal] = true
				updatedEndpoint.Urls[index] = sidecarUrl
			} else if strings.EqualFold(mode, ServerLess) {
				prodEpVal = fmt.Sprintf("%v://%v.%v", protocol, prodEpVal, epNamespace)
				updatedEndpoint.Urls[index] = prodEpVal
			} else {
				prodEpVal = fmt.Sprintf("%v://%v.%v:%v", protocol, prodEpVal, epNamespace, port)
				updatedEndpoint.Urls[index] = prodEpVal
			}
		}
	}
	return nil
}
