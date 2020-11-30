package envoy

import (
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/mitchellh/mapstructure"
	"path"
	"strings"
)
type Endpoints struct {
	Type string   `yaml:"type"`
	Urls []string `yaml:"urls"`
}

// generateFieldsFromSwagger3 using swagger
func swagger2Populate(def *APIDefinition, document *loads.Document) error {
	def.ID.APIName = document.Spec().Info.Title
	def.ID.Version = document.Spec().Info.Version
	def.ID.ProviderName = "admin"
	def.Description = document.Spec().Info.Description
	def.Context = fmt.Sprintf("/%s/%s", def.ID.APIName, def.ID.Version)
	def.ContextTemplate = fmt.Sprintf("/%s/{version}", def.ID.APIName)
	def.Tags = swagger2Tags(document)

	// fill basepath from swagger
	if document.BasePath() != "" {
		def.Context = path.Clean(fmt.Sprintf("/%s/%s", document.BasePath(), def.ID.Version))
		def.ContextTemplate = path.Clean(fmt.Sprintf("/%s/{version}", document.BasePath()))
	}

	// override basepath if wso2 extension provided
	if basepath, ok := swagger2XWO2BasePath(document); ok {
		def.Context = path.Clean(basepath)
		def.ContextTemplate = path.Clean(basepath)
		if !strings.Contains(basepath, "{version}") {
			def.Context = path.Clean(basepath + "/" + def.ID.Version)
			def.ContextTemplate = path.Clean(basepath + "/{version}")
			def.IsDefaultVersion = true
		} else {
			def.ContextTemplate = path.Clean(basepath)
			def.Context = path.Clean(strings.ReplaceAll(basepath, "{version}", def.ID.Version))
		}
	}

	// trim spaces if available
	def.ID.APIName = strings.ReplaceAll(def.ID.APIName, " ", "")
	def.ID.Version = strings.ReplaceAll(def.ID.Version, " ", "")
	def.Context = strings.ReplaceAll(def.Context, " ", "")
	def.ContextTemplate = strings.ReplaceAll(def.ContextTemplate, " ", "")

	cors, ok, err := swagger2XWSO2Cors(document)
	if err != nil && ok {
		return err
	}
	if ok {
		def.CorsConfiguration = cors
	}

	//prodEp, foundProdEp, err := swagger2XWSO2ProductionEndpoints(document)
	//if err != nil && foundProdEp {
	//	return err
	//}
	//sandboxEp, foundSandboxEp, err := swagger2XWSO2SandboxEndpoints(document)
	//if err != nil && foundSandboxEp {
	//	return err
	//}
	//
	//if foundProdEp || foundSandboxEp {
	//	ep, err := BuildAPIMEndpoints(prodEp, sandboxEp)
	//	if err != nil {
	//		return err
	//	}
	//	def.EndpointConfig = &ep
	//}
	return nil
}

func swagger2Tags(document *loads.Document) []string {
	tags := make([]string, len(document.Spec().Tags))
	for i, v := range document.Spec().Tags {
		tags[i] = v.Name
	}
	return tags
}

func swagger2XWO2BasePath(document *loads.Document) (string, bool) {
	if v, ok := document.Spec().Extensions["x-wso2-basePath"]; ok {
		str, ok := v.(string)
		return str, ok
	}
	return "", false
}

func swagger2XWSO2Cors(document *loads.Document) (*CorsConfiguration, bool, error) {
	if v, ok := document.Spec().Extensions["x-wso2-cors"]; ok {
		var cors CorsConfiguration
		err := mapstructure.Decode(v, &cors)
		if err != nil {
			return nil, true, err
		}
		cors.CorsConfigurationEnabled = true
		return &cors, true, nil
	}
	return nil, false, nil
}
//
//func swagger2XWSO2ProductionEndpoints(document *loads.Document) (*Endpoints, bool, error) {
//	if v, ok := document.Spec().Extensions["x-wso2-production-endpoints"]; ok {
//		var prodEp Endpoints
//		err := mapstructure.Decode(v, &prodEp)
//		if err != nil {
//			return nil, true, err
//		}
//		return &prodEp, true, nil
//	}
//	return &Endpoints{}, false, nil
//}
//
//func swagger2XWSO2SandboxEndpoints(document *loads.Document) (*Endpoints, bool, error) {
//	if v, ok := document.Spec().Extensions["x-wso2-sandbox-endpoints"]; ok {
//		var sandboxEp Endpoints
//		err := mapstructure.Decode(v, &sandboxEp)
//		if err != nil {
//			return nil, true, err
//		}
//		return &sandboxEp, true, nil
//	}
//	return &Endpoints{}, false, nil
//}

