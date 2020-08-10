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
	"github.com/getkin/kin-openapi/openapi3"
	v1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/serving/v1alpha1"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
)

var logEp = log.Log.WithName("swagger.endpoints")

const (
	ServerLess = "serverless"
	Sidecar    = "sidecar"
	privateJet = "privateJet"
)

//XMgwProductionEndpoint represents the structure of endpoint
type XMgwProductionEndpoint struct {
	Urls []string `yaml:"urls" json:"urls"`
}

// HandleMgwEndpoints gets endpoint from swagger and replace it with targetendpoint kind service endpoint and
// returns a map of sidecar endpoints
func HandleMgwEndpoints(client *client.Client, swagger *openapi3.Swagger, mode string, apiNamespace string) (
	map[string]string, error) {
	sideCarEndpoints := make(map[string]string) // map endpoint name -> endpoint URL
	// TODO: ~rnk `sideCarEndpoints` can be converted to array of str - no need endpoint URL

	// API level endpoint
	endpointData, checkEndpoint := swagger.Extensions[EndpointExtension]
	if checkEndpoint {
		logEp.Info("API level endpoint is defined")
		endpointJson, checkJsonRaw := endpointData.(json.RawMessage)
		if !checkJsonRaw {
			logEp.Error(errs.New("value is not a json.RawMessage"),
				"Invalid format of Target Endpoint definition in swagger")
		}

		prodEp := XMgwProductionEndpoint{}
		if err := json.Unmarshal(endpointJson, &prodEp); err != nil {
			logEp.Error(err, "Invalid format of Target Endpoint definition in swagger")
		}

		// URL list to update swagger definition
		epUrlsForSwg := make([]string, len(prodEp.Urls))

		for index, prodEpVal := range prodEp.Urls {
			prodEpUrl, errUrl := url.ParseRequestURI(prodEpVal)
			if errUrl == nil { // Target EP is a valid URL
				epUrlsForSwg[index] = prodEpUrl.RequestURI()
			} else { // Target EP is a name of Target EP CR
				epNamespace := apiNamespace // namespace of the endpoint
				if namespacedEp := strings.Split(prodEpVal, "."); len(namespacedEp) == 2 {
					epNamespace = namespacedEp[1]
					prodEpVal = namespacedEp[0]
				}

				targetEpCr := &wso2v1alpha1.TargetEndpoint{} // CR of the Target Endpoint
				erCr := k8s.Get(client, types.NamespacedName{Namespace: epNamespace, Name: prodEpVal}, targetEpCr)
				if erCr != nil {
					return nil, erCr
				}

				protocol := targetEpCr.Spec.ApplicationProtocol
				port := strconv.Itoa(int(targetEpCr.Spec.Ports[0].Port))
				if strings.EqualFold(mode, Sidecar) { // sidecar mode
					sidecarUrl := fmt.Sprintf("%v://localhost:%v", protocol, port)
					sideCarEndpoints[prodEpVal] = sidecarUrl
					epUrlsForSwg[index] = sidecarUrl
				} else if strings.EqualFold(mode, ServerLess) {
					prodEpVal = fmt.Sprintf("%v://%v.%v.svc.cluster.local", protocol, prodEpVal, epNamespace)
					epUrlsForSwg[index] = prodEpVal
				} else {
					prodEpVal = fmt.Sprintf("%v://%v.%v:%v", protocol, prodEpVal, epNamespace, port)
					epUrlsForSwg[index] = prodEpVal
				}
			}
		}

		// update swagger definition
		prodEp.Urls = epUrlsForSwg
		swagger.Extensions[EndpointExtension] = prodEp
	}

	//resource level endpoint
	for pathName, path := range swagger.Paths {
		if path.Get != nil {
			getEp, gcep := path.Get.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, getEp, sideCarEndpoints, gcep, apiNamespace, mode)
			assignGetEps(swagger, eps)
		}
		if path.Post != nil {
			postEp, pocep := path.Post.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, postEp, sideCarEndpoints, pocep, apiNamespace, mode)
			assignPostEps(swagger, eps)
		}
		if path.Put != nil {
			putEp, pucep := path.Put.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, putEp, sideCarEndpoints, pucep, apiNamespace, mode)
			assignPutEps(swagger, eps)
		}
		if path.Delete != nil {
			deleteEp, dcep := path.Delete.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, deleteEp, sideCarEndpoints, dcep, apiNamespace, mode)
			assignDeleteEps(swagger, eps)
		}
		if path.Patch != nil {
			pEp, pAvl := path.Patch.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, sideCarEndpoints, pAvl, apiNamespace, mode)
			assignPatchEps(swagger, eps)
		}
		if path.Head != nil {
			pEp, pAvl := path.Head.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, sideCarEndpoints, pAvl, apiNamespace, mode)
			assignHeadEps(swagger, eps)
		}
		if path.Options != nil {
			pEp, pAvl := path.Options.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, sideCarEndpoints, pAvl, apiNamespace, mode)
			assignOptionsEps(swagger, eps)
		}
	}
	return sideCarEndpoints, nil
}

func assignGetEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Get.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignPutEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Put.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignPostEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Post.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignDeleteEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Delete.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignPatchEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Patch.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignHeadEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Head.Extensions[EndpointExtension] = value
			}
		}
	}
}

func assignOptionsEps(swagger *openapi3.Swagger, resLevelEp map[string]XMgwProductionEndpoint) {
	for pathName, path := range swagger.Paths {
		for mapPath, value := range resLevelEp {
			if strings.EqualFold(pathName, mapPath) {
				path.Options.Extensions[EndpointExtension] = value
			}
		}
	}
}

func resolveEps(client *client.Client, pathName string, resourceGetEp interface{}, endpointNames map[string]string, checkResourceEP bool,
	userNameSpace string, mode string) map[string]XMgwProductionEndpoint {
	var checkr []string
	var resLevelEp = make(map[string]XMgwProductionEndpoint)

	//resourceGetEp, checkResourceEP := path.Get.Extensions[endpointExtension]
	if checkResourceEP {
		prodEp := XMgwProductionEndpoint{}
		var endPoint string
		ResourceEndpointJson, checkJsonResource := resourceGetEp.(json.RawMessage)
		if checkJsonResource {
			err := json.Unmarshal(ResourceEndpointJson, &endPoint)
			if err == nil {

				extractData := strings.Split(endPoint, ".")
				if len(extractData) == 2 {
					userNameSpace = extractData[1]
					endPoint = extractData[0]
				}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				erCr := k8s.Get(client, types.NamespacedName{Namespace: userNameSpace, Name: endPoint}, targetEndpointCr)

				if erCr != nil && errors.IsNotFound(erCr) {
					logEp.Error(err, "targetEndpoint CRD object is not found")
				} else if erCr != nil {
					logEp.Error(err, "Error in getting targetendpoint CRD object")
				}
				if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), ServerLess) {
					currentService := &v1.Service{}
					err = k8s.Get(client, types.NamespacedName{Namespace: userNameSpace,
						Name: endPoint}, currentService)
				} else {
					currentService := &corev1.Service{}
					err = k8s.Get(client, types.NamespacedName{Namespace: userNameSpace,
						Name: endPoint}, currentService)
				}
				if err != nil && errors.IsNotFound(err) && mode != Sidecar {
					logEp.Error(err, "service not found")
				} else if err != nil && mode != Sidecar {
					logEp.Error(err, "Error in getting service")
				} else {
					protocol := targetEndpointCr.Spec.ApplicationProtocol
					if mode == Sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Ports[0].Port))
						endpointNames[targetEndpointCr.Name] = endPointSidecar
						checkr = append(checkr, endPointSidecar)
					}
					if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), ServerLess) {

						endPoint = protocol + "://" + endPoint + "." + userNameSpace + ".svc.cluster.local"
						checkr = append(checkr, endPoint)

					} else {
						endPoint = protocol + "://" + endPoint + "." + userNameSpace + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Ports[0].Port))
						checkr = append(checkr, endPoint)
					}
					prodEp.Urls = checkr
					resLevelEp[pathName] = prodEp
				}
			} else {
				err := json.Unmarshal(ResourceEndpointJson, &prodEp)
				if err == nil {
					lengthOfUrls := len(prodEp.Urls)
					endpointList := make([]string, lengthOfUrls)
					isServiceDef := false
					for index, urlVal := range prodEp.Urls {
						endpointUrl, err := url.Parse(urlVal)
						if err != nil {
							currentService := &corev1.Service{}
							targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
							err = k8s.Get(client, types.NamespacedName{Namespace: userNameSpace,
								Name: urlVal}, currentService)
							erCr := k8s.Get(client, types.NamespacedName{Namespace: userNameSpace, Name: urlVal}, targetEndpointCr)
							if err == nil && erCr == nil || mode == Sidecar {
								endpointNames[urlVal] = urlVal
								protocol := targetEndpointCr.Spec.ApplicationProtocol
								if mode == Sidecar {
									urlValSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Ports[0].Port))
									endpointNames[urlVal] = urlValSidecar
									endpointList[index] = urlValSidecar
								} else {
									urlVal = protocol + "://" + urlVal + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Ports[0].Port))
									endpointList[index] = urlVal
								}
								isServiceDef = true
							}
						} else {
							endpointNames[endpointUrl.Hostname()] = endpointUrl.Hostname()
						}
					}

					if isServiceDef {
						prodEp.Urls = endpointList
						resLevelEp[pathName] = prodEp
					}
				}
			}
		}
	}
	return resLevelEp
}
