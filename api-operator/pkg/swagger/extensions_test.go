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
	"github.com/getkin/kin-openapi/openapi3"
	"testing"
)

func TestApiBasePath(t *testing.T) {

	var err error
	var openapiV3Result *openapi3.Swagger
	var apiBasePath string
	var openapiV3 string

	openapiV3 = readFileContent(t, "../../test/swagger/openapi_v3.yaml")
	openapiV3Result, err = GetSwaggerV3(&openapiV3)

	if err != nil {
		t.Error("error while reading the swagger file")
	}
	apiBasePath = ApiBasePath(openapiV3Result)

	if apiBasePath == "" {
		t.Error("getting the api base path for valid openapi should not return empty")
	}

	openapiV3 = readFileContent(t, "../../test/swagger/openapi_v3_x_base_path.yaml")
	openapiV3Result, err = GetSwaggerV3(&openapiV3)

	if err != nil {
		t.Error("error while reading the swagger file")
	}
	apiBasePath = ApiBasePath(openapiV3Result)

	if apiBasePath == "" {
		t.Error("getting the api base path for valid openapi should not return empty")
	}

	openapiV3 = readFileContent(t, "../../test/swagger/openapi_v3_invalid.yaml")
	openapiV3Result, err = GetSwaggerV3(&openapiV3)

	if err != nil {
		t.Error("error while reading the swagger file")
	}
	apiBasePath = ApiBasePath(openapiV3Result)

	if apiBasePath != "" {
		t.Error("getting the api base path for invalid openapi should return empty")
	}
}