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
	"io/ioutil"
	"testing"
)

func TestGetSwaggerV3ForSwaggerV2(t *testing.T) {

	swaggerV2 := readFileContent(t, "../../test/swagger/swagger_v2.yaml")
	swaggerV2Result, err := GetSwaggerV3(&swaggerV2)

	if err != nil  {
		t.Error("getting the swagger v2 for valid swagger v2 should not return an error")
	}

	if swaggerV2Result == nil {
		t.Error("getting the swagger v2 for valid swagger v2 should not return nil")
	}

}

func TestGetSwaggerV3ForOpenAPIV3(t *testing.T) {

	openapiV3 := readFileContent(t, "../../test/swagger/openapi_v3.yaml")
	openapiV3Result, err := GetSwaggerV3(&openapiV3)

	if err != nil  {
		t.Error("getting the openapi v3 for valid openapi v3 should not return an error")
	}

	if openapiV3Result == nil {
		t.Error("getting the openapi v3 for valid openapi v3 should not return nil")
	}

}

func TestGetSwaggerV3ForInvalidOpenAPIV3(t *testing.T) {

	var err error
	var openapiV3 string
	var openapiV3Result *openapi3.Swagger

	openapiV3 = readFileContent(t, "../../test/swagger/openapi_v3_invalid.yaml")
	openapiV3Result, err = GetSwaggerV3(&openapiV3)

	if err != nil  {
		t.Error("getting the openapi v3 for valid openapi v3 should not return an error")
	}

	if openapiV3Result == nil {
		t.Error("getting the openapi v3 for valid openapi v3 should not return nil")
	}

	openapiV3 = "Invalid OpenAPI"
	openapiV3Result, err = GetSwaggerV3(&openapiV3)

	if err == nil  {
		t.Error("getting the openapi v3 for invalid openapi v3 should return an error")
	}

	if openapiV3Result != nil {
		t.Error("getting the openapi v3 for invalid openapi v3 should return nil")
	}
}

func readFileContent(t *testing.T, path string) string {

	data, err := ioutil.ReadFile(path)

	if err != nil {
		t.Error("error while reading the openapi file")
	}
	return string(data)
}