package swagger

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logSec = log.Log.WithName("swagger.security")

type path struct {
	Security []map[string][]string `json:"security"`
}

func GetSecurityMap(apiSwagger *openapi3.Swagger) (map[string][]string, bool, int, error) {
	//get all the securities defined in swagger
	var securityMap = make(map[string][]string)
	var securityDef path
	//get API level security
	apiLevelSecurity, isDefined := apiSwagger.Extensions[SecurityExtension]
	var APILevelSecurity []map[string][]string
	if isDefined {
		logSec.Info("API level security is defined")
		rawmsg := apiLevelSecurity.(json.RawMessage)
		errsec := json.Unmarshal(rawmsg, &APILevelSecurity)
		if errsec != nil {
			logSec.Error(errsec, "error unmarshaling API level security ")
			return securityMap, isDefined, len(securityDef.Security), errsec
		}
		for _, value := range APILevelSecurity {
			for secName, val := range value {
				securityMap[secName] = val
			}
		}
	} else {
		logSec.Info("API Level security is not defined")
	}
	//get resource level security
	resLevelSecurity, resSecIsDefined := apiSwagger.Extensions[PathsExtension]
	var resSecurityMap map[string]map[string]path

	if resSecIsDefined {
		rawSec := resLevelSecurity.(json.RawMessage)
		errSec := json.Unmarshal(rawSec, &resSecurityMap)
		if errSec != nil {
			logSec.Error(errSec, "error unmarshal into resource level security")
			return securityMap, isDefined, len(securityDef.Security), errSec
		}
		for _, path := range resSecurityMap {
			for _, sec := range path {
				securityDef = sec // TODO: rnk: Issue: If the final resource path not defined security no security for all other resources
			}
		}
	}
	if len(securityDef.Security) > 0 {
		logSec.Info("Resource level security is defined")
		for _, obj := range resSecurityMap {
			for _, obj := range obj {
				for _, value := range obj.Security {
					for secName, val := range value {
						securityMap[secName] = val
					}
				}
			}
		}
	} else {
		logSec.Info("Resource level security is not defined")
	}
	return securityMap, isDefined, len(securityDef.Security), nil
}
