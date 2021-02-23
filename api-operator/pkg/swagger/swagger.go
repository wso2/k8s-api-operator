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
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("swagger")

// GetSwaggerV3 returns the openapi3.Swagger of given swagger string
func GetSwaggerV3(swaggerStr *string) (*openapi3.Swagger, error) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(*swaggerStr))
	if err != nil {
		return GetSwaggerV2(swaggerStr)
	}

	swaggerV3Version := swagger.OpenAPI
	logger.Info("Swagger version", "version", swaggerV3Version)

	if swaggerV3Version != "" {
		return swagger, err
	} else {
		logger.Info("OpenAPI v3 not found. Hence converting Swagger v2 to Swagger v3")
		return GetSwaggerV2(swaggerStr)
	}
}

// GetSwaggerV2 returns the openapi2.Swagger of given swagger string
func GetSwaggerV2(swaggerStr *string) (*openapi3.Swagger, error) {
	var swagger2 openapi2.Swagger
	// ignore error
	_ = json.Unmarshal([]byte(*swaggerStr), &swagger2)
	if swagger2.BasePath == "" {
		if err := yaml.Unmarshal([]byte(*swaggerStr), &swagger2); err != nil {
			return nil, err
		}
	}

	swaggerV3, err2 := openapi2conv.ToV3Swagger(&swagger2)
	return swaggerV3, err2
}
