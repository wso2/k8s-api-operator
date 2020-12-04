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
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logExt = log.Log.WithName("swagger.extensions")

const (
	EndpointExtension           = "x-wso2-endpoints"
	ProductionEndpointExtension = "x-wso2-production-endpoints"
	ApiBasePathExtension        = "x-wso2-basePath"
	DeploymentModeExtension     = "x-wso2-mode"
	VhostExtension              = "x-wso2-vhost"
	SpecExtension               = "x-wso2-spec"
	EngageEnforcerExtension     = "x-wso2-engage-enforcer"
	SecurityExtension           = "security"
	PathsExtension              = "paths"
	SecuritySchemeExtension     = "securitySchemes"
)

func ApiBasePath(swagger *openapi3.Swagger) string {
	var apiBasePath string

	basePathData, checkBasePath := swagger.Extensions[ApiBasePathExtension]
	if checkBasePath {
		basePathJson, checkJsonRaw := basePathData.(json.RawMessage)
		if checkJsonRaw {
			err := json.Unmarshal(basePathJson, &apiBasePath)
			if err != nil {
				logExt.Error(err, "Error unmarshal API base path path")
			}
		} else {
			logExt.Error(nil, "Wrong format of API base path in the swagger")
		}
	} else {
		logExt.Error(nil, "API base path extension not found in the swagger")
	}

	return apiBasePath
}

func EpDeployMode(api *wso2v1alpha2.API, swagger *openapi3.Swagger) (string, error) {
	var epDeployMode string
	numOfSwaggers := len(api.Spec.Definition.SwaggerConfigmapNames)

	if numOfSwaggers > 1 {
		// override mode in swaggers if there are multiple swaggers
		if api.Spec.Mode != "" {
			epDeployMode = api.Spec.Mode.String()
			logExt.Info("Set endpoint deployment mode in multi swagger mode given in API crd", "mode", epDeployMode)
			return epDeployMode, nil
		}

		// if not defined in swagger or CRD mode set default
		logExt.Info("Set endpoint deployment mode in multi swagger mode with default mode", "default_mode", privateJet)
		return privateJet, nil

	} else if numOfSwaggers < 1 {
		err := errors.New("no swagger configmap defined")
		return "", err
	}

	// override 'instance.Spec.Mode' if there is only one swagger
	// get the mode from swagger file
	modeExt, isModeDefined := swagger.Extensions[DeploymentModeExtension]
	if isModeDefined {
		modeRawStr, _ := modeExt.(json.RawMessage)
		err := json.Unmarshal(modeRawStr, &epDeployMode)
		if err != nil {
			logExt.Error(err, "Error unmarshal mode in swagger", "field", DeploymentModeExtension)
			return "", err
		}

		return epDeployMode, nil
	}

	logExt.Info("Deployment mode is not found in the swagger and setting to default", "default_mode", privateJet)
	return privateJet, nil
}
