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

package envoy

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func getFakeClient(obj runtime.Object) *client.Client {

	objs := []runtime.Object{obj}
	s := scheme.Scheme
	cl := fake.NewFakeClientWithScheme(s, objs...)
	return &cl
}

func readFileContent(t *testing.T, path string) string {

	data, err := ioutil.ReadFile(path)

	if err != nil {
		t.Error("error while reading the openapi file")
	}
	return string(data)
}

func TestCreateDirectories(t *testing.T) {

	pwd, _ := os.Getwd()
	directoryName := "/project123"
	fullPath := pwd + directoryName

	err := createDirectories(fullPath)
	if err != nil {
		t.Error("create directory should not return an error")
	}

	err1 := os.RemoveAll(fullPath)
	if err1 != nil {
		t.Error("error while removing the project")
	}

}

func TestJsonToYaml(t *testing.T) {

	dataValues := []byte("value1")
	_, err := jsonToYaml(dataValues)
	if err != nil {
		t.Error("converting from Json to Yaml should not return an error")
	}

}

func TestGetTempFileForSwagger(t *testing.T) {

	_, err := getTempFileForSwagger("Test String", "file.yaml")

	if err != nil {
		t.Error("getting temp file for Swagger (yaml) should not return an error")
	}

	_, err1 := getTempFileForSwagger("Test String", "file.json")

	if err1 != nil {
		t.Error("getting temp file for Swagger (Json) should not return an error")
	}
}

func TestGetSwaggerData(t *testing.T) {

	configMapData := make(map[string]string, 0)
	openapiV3 := readFileContent(t, "../../test/envoy/openapi_v3.yaml")
	configMapData["swagger.yaml"] = openapiV3

	config := k8s.NewConfMap()
	config.Name = "test-cm"
	config.Data = configMapData

	os.Setenv(apiOperatorConfigHome, "../../build/controller_resources")
	zipFile, cleanupFunc, err := getSwaggerData(config)
	defer cleanupFunc()
	if err != nil {
		t.Error("getting swagger data file should not return an error")
	}

	if zipFile == "" {
		t.Error("getting swagger data file should return a proper file name")
	}

}

func TestGetSwaggerDataForInvalidConfigMap(t *testing.T) {

	config := k8s.NewConfMap()
	config.Name = "test-cm"

	zipFile, _, err := getSwaggerData(config)
	if err == nil {
		t.Error("getting swagger data file for invalid config map should return an error")
	}

	if zipFile != "" {
		t.Error("getting swagger data file for invalid config map should not return a proper file name")
	}

}

func TestZipFiles(t *testing.T) {

	var files []string
	pwd, _ := os.Getwd()
	filePath := path.Join(path.Dir(pwd), "../test/envoy/openapi_v3.yaml")
	files = append(files, filePath)

	zipPath := "../../test/envoy/test.zip"

	err := ZipFiles(zipPath, files)
	if err != nil {
		t.Error("getting zip file for valid file paths should not return an error")
	}

	err1 := os.RemoveAll(zipPath)
	if err1 != nil {
		t.Error("error while removing the zip file")
	}

}

func TestGetCert(t *testing.T) {

	var err error
	mgwSecret := k8s.NewSecret()
	secretName := "envoymgw-cert-secret"
	mgwSecret.Name = secretName
	mgwSecret.Namespace = "wso2-system"

	secretData := make(map[string][]byte, 0)
	secretData["cert"] = []byte("sample-value")
	mgwSecret.Data = secretData

	cl := getFakeClient(mgwSecret)
	err = getCert(cl, secretName)

	if err != nil {
		t.Error("getting mgw cert for valid values should not return an error")
	}

	err = getCert(cl, "invalid")

	if err == nil {
		t.Error("getting mgw cert for invalid values should return an error")
	}

}

func TestGetCertForEmptyData(t *testing.T) {

	var err error
	mgwSecret := k8s.NewSecret()
	secretName := "envoymgw-cert-secret"
	mgwSecret.Name = secretName
	mgwSecret.Namespace = "wso2-system"

	cl := getFakeClient(mgwSecret)
	err = getCert(cl, secretName)

	if err == nil {
		t.Error("getting mgw cert for empty data should return an error")
	}

}

func TestGetZipData(t *testing.T) {

	config := k8s.NewConfMap()
	config.Name = "test-cm"
	cmData := make(map[string][]byte, 0)
	cmData["api"] = []byte("sample-value")
	config.BinaryData = cmData

	fileName, err := getZipData(config)

	if err != nil {
		t.Error("getting zip data for valid config map should not return an error")
	}

	if fileName == "" {
		t.Error("getting zip data for valid config map should return an empty value")
	}
}

func TestGetZipDataInvalidConfigMap(t *testing.T) {

	config := k8s.NewConfMap()
	config.Name = "test-cm"
	fileName, err := getZipData(config)

	if err == nil {
		t.Error("getting zip data for invalid config map should return an error")
	}

	if fileName != "" {
		t.Error("getting zip data for invalid config map should not return an empty value")
	}
}
