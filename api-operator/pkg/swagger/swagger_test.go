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
	"path/filepath"
	"testing"
)

func TestOrderPaths(t *testing.T) {
	paths := []string{"/products/*", "/tv/", "/products/tv", "/", "/products", "/*"}
	want := []string{"/products/tv", "/products", "/tv/", "/", "/products/*", "/*"}
	orderPaths(paths)

	isErr := false
	for i := range paths {
		if paths[i] != want[i] {
			isErr = true
		}
	}
	if isErr {
		t.Errorf("Ordered paths: %v, want: %v", paths, want)
	}
}

func TestPrettyStringOrderedByPath(t *testing.T) {
	swg, err := readJSONResourceFile("test_resources/sample-swagger.json")
	if err != nil {
		t.Fatal("Error reading sample swagger definition")
	}
	prettySwg := PrettyStringOrderedByPath(swg)

	want, err := readResource("test_resources/prettified-path-ordered-swagger.json")
	if err != nil {
		t.Fatal("Error reading sample resource")
	}
	s := string(want)
	if prettySwg != s {
		t.Error("Prettified json is not ordered with swagger paths")
	}
}

func readJSONResourceFile(path string) (*openapi3.Swagger, error) {
	bytes, err := readResource(path)
	if err != nil {
		return nil, err
	}
	s := string(bytes)

	return GetSwaggerV3(&s)
}

func readResource(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.FromSlash(path))
}
