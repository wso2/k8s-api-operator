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

package registry

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/registry/utils"
	"strings"
)

const HTTPS Type = "HTTPS"

// Copy of Docker Hub configs as HTTPS private registry configs
var httpsReg = *dockerHub

func getHttpsRegConfigFunc(repoName string, imgName string, tag string) *Config {
	httpsReg.ImagePath = fmt.Sprintf("%s/%s:%s", repoName, imgName, tag)
	httpsReg.IsImageExist = func(config *Config, auth utils.RegAuth, image string, tag string) (b bool, err error) {
		logger.Info("Checking for image in HTTPS registry", "Registry URL", auth.RegistryUrl)
		return utils.IsImageExists(auth, getImageWithoutReg(image), tag)
	}
	return &httpsReg
}

// getImageWithoutReg remove registry host name if it exists in the image name
func getImageWithoutReg(image string) string {
	splits := strings.Split(image, "/")
	switch len(splits) {
	case 3: // image in format "my-reg-host:5000/operator-demo/pets:v1"
		return fmt.Sprintf("%s/%s", splits[1], splits[2])
	case 2: // image in format "my-reg-host:5000/pets:v1"
		return splits[1]
	default:
		return image
	}
}

func init() {
	addRegistryConfig(HTTPS, getHttpsRegConfigFunc)
}
