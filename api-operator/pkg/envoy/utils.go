package envoy

import (
	"crypto/x509"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apim"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/utils"
	yaml2 "gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// directories to be created
var dirs = []string{
	"Meta-information",
	"Image",
	"Docs",
	"Docs/FileContents",
	"Sequences",
	"Sequences/fault-sequence",
	"Sequences/in-sequence",
	"Sequences/out-sequence",
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

	if _, err := swaggerFile.Write([]byte(swaggerData)); err != nil {
		logDeploy.Error(err, "Error while writing to temp swagger file")
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

func getZipData (config *corev1.ConfigMap) (string, error){
	file, err := ioutil.TempFile("", "api-binary.*.zip")
	if err != nil {
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

func getSwaggerData (config *corev1.ConfigMap) (string, func(), error){
	swaggerFileName, errSwagger := maps.OneKey(config.Data)
	if errSwagger != nil {
		logDeploy.Error(errSwagger, "Error in the swagger configMap data", "data", config.Data)
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
	def := &apim.APIDefinition{}

	err = getAPIData(def, doc)
	if err != nil {
		return "", nil, err
	}
	if def.EndpointConfig != nil {
		def.ProductionUrl = ""
		def.SandboxUrl = ""
	}
	apiData, err := yaml2.Marshal(def)
	if err != nil {
		return "", nil, err
	}
	// convert and save swagger as yaml
	yamlSwagger, err := jsonToYaml(doc.Raw())
	if err != nil {
		return "", nil, err
	}

	swaggerDirectory, _ := ioutil.TempDir("", "api-swagger-dir*")
	apiYamlPath := filepath.Join(swaggerDirectory, filepath.FromSlash("Meta-information/api.yaml"))
	swaggerSavePath := filepath.Join(swaggerDirectory, filepath.FromSlash("Meta-information/swagger.yaml"))
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
	swaggerZipFile, err, cleanupFunc := utils.CreateZipFileFromProject(swaggerDirectory, true)
	if err != nil {
		return "", nil, err
	}
	return swaggerZipFile, cleanupFunc, nil
}
