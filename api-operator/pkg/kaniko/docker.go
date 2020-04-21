package kaniko

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/volume"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"text/template"
)

var logDocker = log.Log.WithName("kaniko.docker")

const (
	truststoreSecretName     = "truststorepass"
	dockerFileTemplate       = "dockerfile-template"
	wso2NameSpaceConst       = "wso2-system"
	encodedTrustsorePassword = "YmFsbGVyaW5h"
	truststoreSecretData     = "password"
	dockerFile               = "dockerfile"
	dockerFileLocation       = "/usr/wso2/dockerfile/"
)

// DockerfileProperties represents the type for properties of docker file
type DockerfileProperties struct {
	CertFound             bool
	TruststorePassword    string
	Certs                 map[string]string
	ToolkitImage          string
	RuntimeImage          string
	BalInterceptorsFound  bool
	JavaInterceptorsFound bool
}

// DocFileProp represents the properties of docker file
var DocFileProp = initDocFileProp

// initDocFileProp represents the initial values for DocFileProp
var initDocFileProp = &DockerfileProperties{
	CertFound:             false,
	TruststorePassword:    "",
	Certs:                 map[string]string{},
	ToolkitImage:          "",
	RuntimeImage:          "",
	BalInterceptorsFound:  false,
	JavaInterceptorsFound: false,
}

func InitDocFileProp() {
	DocFileProp = initDocFileProp
}

// HandleDockerFile render the docker file for Kaniko job and add volumes to the Kaniko job
func HandleDockerFile(client *client.Client, userNamespace, apiName string, owner *[]metav1.OwnerReference) error {
	// get docker file template from system namespace
	dockerFileConfMap := k8s.NewConfMap()
	err := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: dockerFileTemplate}, dockerFileConfMap)
	if err != nil {
		logDocker.Error(err, "Error retrieving docker template configmap", "configmap", dockerFileTemplate)
		return err
	}

	// get file name in configmap
	fileName, err := maps.OneKey(dockerFileConfMap.Data)
	if err != nil {
		logDocker.Error(err, "Error retrieving docker template data", "configmap_data", dockerFileConfMap.Data)
		return err
	}

	// set truststore password
	if err := setTruststorePassword(client); err != nil {
		return err
	}

	// get rendered docker file
	renderedDocFile, err := renderedDockerFile(dockerFileConfMap.Data[fileName])
	if err != nil {
		return err
	}

	// final configmap is the configmap that contains the rendered docker file
	finalConfMapName := fmt.Sprintf("%s-%s", apiName, dockerFile)
	dockerDataMap := map[string]string{"Dockerfile": renderedDocFile}
	finalConfMap := k8s.NewConfMapWith(types.NamespacedName{Namespace: userNamespace, Name: finalConfMapName}, &dockerDataMap, nil, owner)
	err = k8s.Apply(client, finalConfMap)
	if err != nil {
		return err
	}

	// add to job volumes
	vol, mount := volume.ConfigMapVolume(apiName+"-"+dockerFile, dockerFileLocation)
	volume.AddVolume(vol, mount)

	return nil
}

// renderedDockerFile returns the rendered docker file using the properties in DocFileProp
func renderedDockerFile(docFileText string) (string, error) {
	docFileTemplate, err := template.New("").Parse(docFileText)
	if err != nil {
		logDocker.Error(err, "Error in generating template with docker file")
		return "", err
	}

	strBuilder := &strings.Builder{}
	err = docFileTemplate.Execute(strBuilder, *DocFileProp)
	if err != nil {
		logDocker.Error(err, "Error rendering dockerfile from template", "template", docFileText, "properties", *DocFileProp)
		return "", err
	}

	return strBuilder.String(), nil
}

// setTruststorePassword sets the truststore password in docker file properties DocFileProp
func setTruststorePassword(client *client.Client) error {
	// get secret if available
	secret := k8s.NewSecret()
	err := k8s.Get(client, types.NamespacedName{Name: truststoreSecretName, Namespace: wso2NameSpaceConst}, secret)
	if err != nil && errors.IsNotFound(err) {
		encodedPw := encodedTrustsorePassword
		decodedPw, err := b64.StdEncoding.DecodeString(encodedPw)
		if err != nil {
			logDocker.Error(err, "Error decoding truststore password")
			return err
		}
		password := string(decodedPw)

		logDocker.Info("Creating a new secret for truststore password")
		trustStoreSecret := k8s.NewSecretWith(types.NamespacedName{
			Namespace: wso2NameSpaceConst,
			Name:      truststoreSecretName,
		}, &map[string][]byte{
			truststoreSecretData: []byte(encodedPw),
		}, nil, nil)

		errSecret := k8s.Create(client, trustStoreSecret)
		logDocker.Info("Error in creating truststore password and ignore it", "error", errSecret)

		DocFileProp.TruststorePassword = password
		return nil
	}
	//get password from the secret
	encodedPw := string(secret.Data[truststoreSecretData])
	decodedPw, err := b64.StdEncoding.DecodeString(encodedPw)
	if err != nil {
		logDocker.Error(err, "Error decoding truststore password")
	}

	DocFileProp.TruststorePassword = string(decodedPw)
	return nil
}
