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
