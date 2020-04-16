package swagger

import (
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("swagger")

// GetSwaggerV3 retuns the openapi3.Swagger of given swagger string
func GetSwaggerV3(swaggerStr *string) (*openapi3.Swagger, error) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(*swaggerStr))
	if err != nil {
		logger.Error(err, "Error loading swagger")
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
