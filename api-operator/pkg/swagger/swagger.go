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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sort"
	"strings"
)

var logger = log.Log.WithName("swagger")

// GetSwaggerV3 returns the openapi3.Swagger of given swagger string
func GetSwaggerV3(swaggerStr *string) (*openapi3.Swagger, error) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(*swaggerStr))
	if err != nil {
		logger.Error(err, "Error loading swagger")
		return nil, err
	}

	swaggerV3Version := swagger.OpenAPI
	logger.Info("Swagger version", "version", swaggerV3Version)

	if swaggerV3Version != "" {
		return swagger, err
	} else {
		logger.Info("OpenAPI v3 not found. Hence converting Swagger v2 to Swagger v3")
		var swagger2 openapi2.Swagger
		err2 := yaml.Unmarshal([]byte(*swaggerStr), &swagger2)
		swaggerV3, err2 := openapi2conv.ToV3Swagger(&swagger2)
		return swaggerV3, err2
	}
}

func PrettyString(swagger *openapi3.Swagger) string {
	marshal, err := swagger.MarshalJSON()
	if err != nil {
		logger.Error(err, "Error marshalling swagger")
	}
	prettyJSON, err := prettifyJSON(marshal)
	if err != nil {
		logger.Error(err, "Error prettifying swagger JSON")
	}
	return string(prettyJSON)
}

func PrettyStringOrderedByPath(swagger *openapi3.Swagger) string {
	const emptyPath = ",\"paths\":{\"xxx\":{}}"
	var emptySample = openapi3.Paths{"xxx": {}}
	const emptyPathFmt = ",\"paths\":{%v}"

	paths := swagger.Paths
	pathNames := make([]string, 0, len(paths))
	for path := range paths {
		pathNames = append(pathNames, path)
	}
	orderPaths(pathNames)
	pathsStr := make([]string, 0, len(paths))
	for _, name := range pathNames {
		marshal, err := json.Marshal(paths[name])
		if err != nil {
			logger.Error(err, "Error marshalling json path", "path", name, "path_item", paths[name])
		}
		pathsStr = append(pathsStr, fmt.Sprintf("\"%v\": %v", name, string(marshal)))
	}
	orderedPathsStr := strings.Join(pathsStr, ",")

	swagger.Paths = emptySample
	m, err := swagger.MarshalJSON()
	if err != nil {
		logger.Error(err, "Error marshalling swagger")
	}
	swaggerSplits := strings.Split(string(m), emptyPath)

	var final string
	if len(swaggerSplits) == 2 {
		final = swaggerSplits[0] + fmt.Sprintf(emptyPathFmt, orderedPathsStr) + swaggerSplits[1]
	} else {
		logger.Error(nil, "Reserved text is used in swagger definition", "reserved_text", emptyPath)
	}

	prettyJSON, err := prettifyJSON([]byte(final))
	if err != nil {
		logger.Error(err, "Error prettifying swagger JSON")
	}
	return string(prettyJSON)
}

func prettifyJSON(b []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, b, "", "  "); err != nil {
		return nil, err
	}
	return prettyJSON.Bytes(), nil
}

func orderPaths(paths []string) {
	const prefixed = "*"
	const separator = "/"
	sort.SliceStable(paths, func(i, j int) bool {
		ips := strings.Split(strings.TrimSuffix(paths[i], separator), separator)
		jps := strings.Split(strings.TrimSuffix(paths[j], separator), separator)
		ip := ips[len(ips)-1]
		jp := jps[len(jps)-1]

		// /products/* vs /products
		if ip == prefixed && jp != prefixed {
			return false
		}
		// /products vs /products/*
		if ip != prefixed && jp == prefixed {
			return true
		}

		// /products vs /orders
		if len(ips) == len(jps) {
			return paths[i] < paths[j]
		}

		// /products vs /products/tv
		return len(ips) > len(jps)
	})
}
