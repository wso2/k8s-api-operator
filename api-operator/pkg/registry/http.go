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

const HTTP Type = "HTTP"

// Copy of Docker Hub configs as HTTP private registry configs
var httpReg = *dockerHub

func getHttpRegConfigFunc(repoName string, imgName string, tag string) *Config {
	httpReg.ImagePath = fmt.Sprintf("%s/%s:%s", repoName, imgName, tag)
	// Add --insecure arg
	// This will not effect Docker Hub configs since this is a copy of it
	httpReg.Args = []string{"--insecure"}
	// replace https with http and use default function
	httpReg.IsImageExist = func(config *Config, auth utils.RegAuth, image string, tag string) (b bool, err error) {
		auth.RegistryUrl = strings.Replace(auth.RegistryUrl, "https://", "http://", 1)
		logger.Info("Checking for image in HTTP registry", "Registry URL", auth.RegistryUrl)
		return utils.IsImageExists(auth, image, imageTag)
	}

	return &httpReg
}

func init() {
	addRegistryConfig(HTTP, getHttpRegConfigFunc)
}
