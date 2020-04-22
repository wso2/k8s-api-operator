package swagger

import (
	"encoding/json"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
)

const (
	privateJet = "privateJet"
)

func ApiBasePath(swagger *openapi3.Swagger) string {
	var apiBasePath string

	basePathData, checkBasePath := swagger.Extensions[ApiBasePathExtension]
	if checkBasePath {
		basePathJson, checkJsonRaw := basePathData.(json.RawMessage)
		if checkJsonRaw {
			err := json.Unmarshal(basePathJson, &apiBasePath)
			if err != nil {
				logger.Error(err, "Error unmarshal API base path path")
			}
		} else {
			logger.Error(nil, "Wrong format of API base path in the swagger")
		}
	} else {
		logger.Error(nil, "API base path extension not found in the swagger")
	}

	return apiBasePath
}

func EpDeployMode(api *wso2v1alpha1.API, swagger *openapi3.Swagger) (string, error) {
	var epDeployMode string
	numOfSwaggers := len(api.Spec.Definition.SwaggerConfigmapNames)

	if numOfSwaggers > 1 {
		// override mode in swaggers if there are multiple swaggers
		if api.Spec.Mode != "" {
			epDeployMode = api.Spec.Mode.String()
			logger.Info("Set endpoint deployment mode in multi swagger mode given in API crd", "mode", epDeployMode)
			return epDeployMode, nil
		}

		// if not defined in swagger or CRD mode set default
		logger.Info("Set endpoint deployment mode in multi swagger mode with default mode", "default_mode", privateJet)
		return privateJet, nil

	} else if numOfSwaggers < 1 {
		err := errors.New("no swagger configmap defined")
		return "", err
	}

	// override 'instance.Spec.Mode' if there is only one swagger
	// get the mode from swagger file
	modeExt, isModeDefined := swagger.Extensions[DeploymentModeExtension]
	if isModeDefined {
		modeRawStr, _ := modeExt.(json.RawMessage)
		err := json.Unmarshal(modeRawStr, &epDeployMode)
		if err != nil {
			logger.Error(err, "Error unmarshal mode in swagger", "field", DeploymentModeExtension)
			return "", err
		}

		return epDeployMode, nil
	}

	logger.Info("Deployment mode is not found in the swagger and setting to default", "default_mode", privateJet)
	return privateJet, nil
}
