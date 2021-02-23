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
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logExt = log.Log.WithName("swagger.extensions")

const (

	ApiBasePathExtension        = "x-wso2-basePath"
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
		if len(swagger.Servers) > 0 {
			splits := strings.Split(swagger.Servers[0].URL, "/")
			return splits[len(splits)-1]
		}
		logExt.Error(nil, "API base path extension not found in the swagger")
	}

	return apiBasePath
}
