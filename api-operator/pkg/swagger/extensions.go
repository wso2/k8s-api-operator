package swagger

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
)

func GetApiBasePath(swagger *openapi3.Swagger) string {
	var apiBasePath string

	basePathData, checkBasePath := swagger.Extensions[ApiBasePathExtention]
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

func GetMode(swagger *openapi3.Swagger) string {
	var mode string
	modeExt, isModeDefined := swagger.Extensions[DeploymentMode]
	if isModeDefined {
		modeRawStr, _ := modeExt.(json.RawMessage)
		err := json.Unmarshal(modeRawStr, &mode)
		if err != nil {
			logger.Error(err, "Error unmarshal mode in swagger", "field", DeploymentMode)
		}
	} else {
		logger.Info("Deployment mode is not set in the swagger", "field", DeploymentMode)
	}

	return mode
}
