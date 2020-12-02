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
)

const QUAY Type = "QUAY"

// getQuayRegConfigFunc Copies Docker Hub configs as Quay registry configs
func getQuayRegConfigFunc(repoName string, imgName string, tag string) *Config {
	var quayReg = getDockerHubConfigFunc(repoName, imgName, tag)
	quayReg.ImagePath = fmt.Sprintf("%s/%s:%s", repoName, imgName, tag)
	return quayReg
}

func init() {
	addRegistryConfig(QUAY, getQuayRegConfigFunc)
}
