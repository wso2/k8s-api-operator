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
	"archive/zip"
	"crypto/x509"
	"encoding/base64"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/utils"
	v2 "github.com/wso2/product-apim-tooling/import-export-cli/specs/v2"
	yaml2 "gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logUtil = log.Log.WithName("mgw.envoy.util")

// directories to be created
var dirs = []string{
	"Definitions",
	"Image",
	"Docs",
	"Docs/FileContents",
	"Sequences",
	"Sequences/fault-sequence",
	"Sequences/in-sequence",
	"Sequences/out-sequence",
	"Client-certificates",
	"Endpoint-certificates",
	"Interceptors",
	"libs",
}
var certPool *x509.CertPool

// createDirectories will create dirs in current working directory
func createDirectories(name string) error {
	for _, dir := range dirs {
		dirPath := filepath.Join(name, filepath.FromSlash(dir))
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// JsonToYaml converts a json string to yaml
func jsonToYaml(jsonData []byte) ([]byte, error) {
	return yaml.JSONToYAML(jsonData)
}

// Get a temp file with swagger
func getTempFileForSwagger(swaggerData string, swaggerFileName string) (*os.File, error) {
	var swaggerFile *os.File
	if strings.Contains(swaggerFileName, "yaml") {
		swaggerFile, _ = ioutil.TempFile("", "api-swagger*.yaml")
	} else {
		swaggerFile, _ = ioutil.TempFile("", "api-swagger*.json")
	}
	if err := os.Chmod(swaggerFile.Name(), 0777); err != nil {
		return nil, err
	}
	if _, err := swaggerFile.Write([]byte(swaggerData)); err != nil {
		logUtil.Error(err, "Error while writing to temp swagger file")
		return nil, err
	}
	swaggerFile.Close()

	return swaggerFile, nil
}

// loads swagger from swaggerDoc
// swagger2.0/OpenAPI3.0 specs are supported
func loadSwagger(swaggerDoc string) (*loads.Document, error) {
	return loads.Spec(swaggerDoc)
}

// loadDefaultSpec loads the API definition
func loadDefaultSpec() (*v2.APIDefinitionFile, error) {

	var apiOperatorConfigHomeDir string
	apiOperatorConfigHomeDir = os.Getenv(apiOperatorConfigHome)

	if apiOperatorConfigHomeDir == "" {
		apiOperatorConfigHomeDir = apiOperatorDefaultConfigHome
	}

	configFileName := "default_api.yaml"
	configFilePath := apiOperatorConfigHomeDir + string(os.PathSeparator) + configFileName

	defaultData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	def := &v2.APIDefinitionFile{}
	marshalErr := yaml.Unmarshal(defaultData, &def)

	if marshalErr != nil {
		return nil, err
	}
	return def, nil
}

// getCert gets the public cert of Envoy MGW Adapter when skip verification is false
func getCert(client *client.Client, mgwCertSecretConf string) error {
	envoyMgwCertSecret := k8s.NewSecret()
	errCert := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace, Name: mgwCertSecretConf},
		envoyMgwCertSecret)
	if errCert != nil {
		return errCert
	}

	certName, errCert := maps.OneKey(envoyMgwCertSecret.Data)
	if errCert != nil {
		return errCert
	}
	cert := string(envoyMgwCertSecret.Data[certName])
	certData := []byte(cert)
	certPool = x509.NewCertPool()
	certPool.AppendCertsFromPEM(certData)

	return nil
}

func getZipData(config *corev1.ConfigMap) (string, error) {
	file, err := ioutil.TempFile("", "api-binary.*.zip")
	if err != nil {
		return "", err
	}
	if err := os.Chmod(file.Name(), 0777); err != nil {
		return "", err
	}
	zipFileName, errZip := maps.OneKey(config.BinaryData)
	if errZip != nil {
		return "", errZip
	}
	zippedData := config.BinaryData[zipFileName]
	if _, err := file.Write(zippedData); err != nil {
		return "", err
	}
	err = file.Close()
	return file.Name(), nil
}

// getMgAdapterSecret Gets the envoymgw-adapter-secret
func getMgAdapterSecret(client *client.Client, secretName string) (*corev1.Secret, error) {

	envoyMgwSecret := k8s.NewSecret()
	errEnvoyMgwSecret := k8s.Get(client, types.NamespacedName{Namespace: config.SystemNamespace,
		Name: secretName}, envoyMgwSecret)

	if errEnvoyMgwSecret != nil {
		return nil, errEnvoyMgwSecret
	}

	return envoyMgwSecret, nil
}

// getAuthToken Get auth token from the secret
func getAuthToken(secret *corev1.Secret) string {

	username := string(secret.Data[usernameProperty])
	password := string(secret.Data[passwordProperty])
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func getSwaggerData(config *corev1.ConfigMap) (string, func(), error) {
	swaggerFileName, errSwagger := maps.OneKey(config.Data)
	if errSwagger != nil {
		logUtil.Error(errSwagger, "Error in the swagger configMap data", "data", config.Data)
		return "", nil, errSwagger
	}
	swaggerData := config.Data[swaggerFileName]

	swaggerFile, errSwaggerFile := getTempFileForSwagger(swaggerData, swaggerFileName)
	if errSwaggerFile != nil {
		return "", nil, errSwaggerFile
	}
	doc, err := loadSwagger(swaggerFile.Name())
	if err != nil {
		return "", nil, err
	}

	definitionFile, err := loadDefaultSpec()
	if err != nil {
		return "", nil, err
	}

	def := &definitionFile.Data

	err = v2.Swagger2Populate(def, doc)
	if err != nil {
		return "", nil, err
	}

	apiData, err := yaml2.Marshal(definitionFile)
	if err != nil {
		return "", nil, err
	}
	// convert and save swagger as yaml
	yamlSwagger, err := jsonToYaml(doc.Raw())
	if err != nil {
		return "", nil, err
	}

	swaggerDirectory, _ := ioutil.TempDir("", "api-swagger-dir*")
	apiYamlPath := filepath.Join(swaggerDirectory, filepath.FromSlash("api.yaml"))
	swaggerSavePath := filepath.Join(swaggerDirectory, filepath.FromSlash("Definitions/swagger.yaml"))
	errCreateDirectory := createDirectories(swaggerDirectory)
	if errCreateDirectory != nil {
		return "", nil, errCreateDirectory
	}
	errWrite := ioutil.WriteFile(swaggerSavePath, yamlSwagger, os.ModePerm)
	if errWrite != nil {
		return "", nil, errWrite
	}
	err = ioutil.WriteFile(apiYamlPath, apiData, os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	swaggerZipFile, err, cleanupFunc := utils.CreateZipFileFromProject(swaggerDirectory, false)
	if err != nil {
		return "", nil, err
	}
	return swaggerZipFile, cleanupFunc, nil
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {
	newZipFile, err1 := os.Create(filename)
	if err1 != nil {
		return err1
	}
	if err := os.Chmod(newZipFile.Name(), 0777); err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err := AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
