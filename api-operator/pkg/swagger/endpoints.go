package swagger

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	v1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/serving/v1alpha1"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

//XMgwProductionEndpoint represents the structure of endpoint
type XMgwProductionEndpoint struct {
	Urls []string `yaml:"urls" json:"urls"`
}

// HandleMgwEndpoints gets endpoint from swagger and replace it with targetendpoint kind service endpoint
func HandleMgwEndpoints(client *client.Client, swagger *openapi3.Swagger, mode string, apiNamespace string) map[string]string {
	endpointNames := make(map[string]string)
	var checkt []string
	//api level endpoint
	endpointData, checkEndpoint := swagger.Extensions[EndpointExtension]
	if checkEndpoint {
		prodEp := XMgwProductionEndpoint{}
		var endPoint string
		endpointJson, checkJsonRaw := endpointData.(json.RawMessage)
		if checkJsonRaw {
			err := json.Unmarshal(endpointJson, &endPoint)
			if err == nil {
				logger.Info("Parsing endpoints and not available root service endpoint")
				//check if service & targetendpoint cr object are available
				extractData := strings.Split(endPoint, ".")
				if len(extractData) == 2 {
					apiNamespace = extractData[1]
					endPoint = extractData[0]
				}
				targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
				erCr := k8s.Get(client, types.NamespacedName{Namespace: apiNamespace, Name: endPoint}, targetEndpointCr)

				if erCr != nil && errors.IsNotFound(erCr) {
					logger.Error(err, "targetEndpoint CRD object is not found")
				} else if erCr != nil {
					logger.Error(err, "Error in getting targetendpoint CRD object")
				}

				if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), ServerLess) {
					currentService := &v1.Service{}
					err = k8s.Get(client, types.NamespacedName{Namespace: apiNamespace,
						Name: endPoint}, currentService)
				} else {
					currentService := &corev1.Service{}
					err = k8s.Get(client, types.NamespacedName{Namespace: apiNamespace,
						Name: endPoint}, currentService)
				}
				if err != nil && errors.IsNotFound(err) && mode != Sidecar {
					logger.Error(err, "service not found")
				} else if err != nil && mode != Sidecar {
					logger.Error(err, "Error in getting service")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					if mode == Sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						endpointNames[targetEndpointCr.Name] = endPointSidecar
						checkt = append(checkt, endPointSidecar)
					}
					if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), ServerLess) {

						endPoint = protocol + "://" + endPoint + "." + apiNamespace + ".svc.cluster.local"
						checkt = append(checkt, endPoint)

					} else {
						endPoint = protocol + "://" + endPoint + "." + apiNamespace + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						checkt = append(checkt, endPoint)
					}
					prodEp.Urls = checkt
					swagger.Extensions[EndpointExtension] = prodEp
				}
			} else {
				err := json.Unmarshal(endpointJson, &prodEp)
				if err == nil {
					lengthOfUrls := len(prodEp.Urls)
					endpointList := make([]string, lengthOfUrls)
					isServiceDef := false
					for index, urlVal := range prodEp.Urls {
						endpointUrl, err := url.Parse(urlVal)
						if err != nil {
							currentService := &corev1.Service{}
							targetEndpointCr := &wso2v1alpha1.TargetEndpoint{}
							err = k8s.Get(client, types.NamespacedName{Namespace: apiNamespace,
								Name: urlVal}, currentService)
							erCr := k8s.Get(client, types.NamespacedName{Namespace: apiNamespace, Name: urlVal}, targetEndpointCr)
							if err == nil && erCr == nil {
								protocol := targetEndpointCr.Spec.Protocol
								urlVal = protocol + "://" + urlVal + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
								if mode == Sidecar {
									urlValSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
									endpointNames[urlVal] = urlValSidecar
									endpointList[index] = urlValSidecar
								} else {
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
						swagger.Extensions[EndpointExtension] = prodEp
					}
				} else {
					logger.Info("error unmarshal endpoint")
				}
			}
		}
	}

	//resource level endpoint
	for pathName, path := range swagger.Paths {
		if path.Get != nil {
			getEp, gcep := path.Get.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, getEp, endpointNames, gcep, apiNamespace, mode)
			assignGetEps(swagger, eps)
		}
		if path.Post != nil {
			postEp, pocep := path.Post.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, postEp, endpointNames, pocep, apiNamespace, mode)
			assignPostEps(swagger, eps)
		}
		if path.Put != nil {
			putEp, pucep := path.Put.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, putEp, endpointNames, pucep, apiNamespace, mode)
			assignPutEps(swagger, eps)
		}
		if path.Delete != nil {
			deleteEp, dcep := path.Delete.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, deleteEp, endpointNames, dcep, apiNamespace, mode)
			assignDeleteEps(swagger, eps)
		}
		if path.Patch != nil {
			pEp, pAvl := path.Patch.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, endpointNames, pAvl, apiNamespace, mode)
			assignPatchEps(swagger, eps)
		}
		if path.Head != nil {
			pEp, pAvl := path.Head.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, endpointNames, pAvl, apiNamespace, mode)
			assignHeadEps(swagger, eps)
		}
		if path.Options != nil {
			pEp, pAvl := path.Options.Extensions[EndpointExtension]
			eps := resolveEps(client, pathName, pEp, endpointNames, pAvl, apiNamespace, mode)
			assignOptionsEps(swagger, eps)
		}
	}
	return endpointNames
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
					logger.Error(err, "targetEndpoint CRD object is not found")
				} else if erCr != nil {
					logger.Error(err, "Error in getting targetendpoint CRD object")
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
					logger.Error(err, "service not found")
				} else if err != nil && mode != Sidecar {
					logger.Error(err, "Error in getting service")
				} else {
					protocol := targetEndpointCr.Spec.Protocol
					if mode == Sidecar {
						endPointSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
						endpointNames[targetEndpointCr.Name] = endPointSidecar
						checkr = append(checkr, endPointSidecar)
					}
					if strings.EqualFold(targetEndpointCr.Spec.Mode.String(), ServerLess) {

						endPoint = protocol + "://" + endPoint + "." + userNameSpace + ".svc.cluster.local"
						checkr = append(checkr, endPoint)

					} else {
						endPoint = protocol + "://" + endPoint + "." + userNameSpace + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
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
								protocol := targetEndpointCr.Spec.Protocol
								if mode == Sidecar {
									urlValSidecar := protocol + "://" + "localhost:" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
									endpointNames[urlVal] = urlValSidecar
									endpointList[index] = urlValSidecar
								} else {
									urlVal = protocol + "://" + urlVal + ":" + strconv.Itoa(int(targetEndpointCr.Spec.Port))
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
