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

package utils

import (
	"fmt"
	"net/url"

	registryclient "github.com/heroku/docker-registry-client/registry"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("registry.utils")

type RegAuth struct {
	RegistryUrl string
	Username    string
	Password    string
}

func IsImageExists(auth RegAuth, imageRepository string, imageName string, tag string) (bool, error) {
	image := fmt.Sprintf("%s/%s", imageRepository, imageName)
	hub, err := registryclient.New(auth.RegistryUrl, auth.Username, auth.Password)
	if err != nil {
		logger.Error(err, "Error connecting to the docker registry", "registry-url", auth.RegistryUrl)
		return false, err
	}

	tags, errRepo := hub.Tags(image)
	if errRepo != nil {
		if errRepo.(*url.Error).Err.(*registryclient.HttpStatusError).Response.StatusCode == 404 {
			logger.Info("Docker repository not found in the registry",
				"registry-url", auth.RegistryUrl, "repository", image)
			return false, nil
		}
		logger.Error(errRepo, "Error getting tags from the image in the docker registry",
			"registry-url", auth.RegistryUrl, "image", image)
		return false, errRepo
	}
	for _, foundTag := range tags {
		if foundTag == tag {
			logger.Info("Found the image tag from the registry", "image", image, "tag", foundTag)
			return true, nil
		}
	}
	return false, nil
}
