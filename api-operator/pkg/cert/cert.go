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

package cert

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/maps"
	corev1 "k8s.io/api/core/v1"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	Path = "/usr/wso2/certs/"
)

var loggerCert = log.Log.WithName("cert")

// AddFromOneKeySecret add the cert to kaniko pod from a secret with only one key
func AddFromOneKeySecret(dockerFileProp *kaniko.DockerFileProperties, certSecret *corev1.Secret, aliasPrefix string) string {
	// add to cert list
	alias := fmt.Sprintf("%s-%s", certSecret.Name, aliasPrefix)
	fileName, err := maps.OneKey(certSecret.Data)
	if err != nil {
		loggerCert.Error(err, "Error reading one key secret. Ignore importing certificate", "secret", certSecret)
		return ""
	}
	Add(dockerFileProp, alias, certSecret.Name, fileName)
	return alias
}

func Add(dockerFileProp *kaniko.DockerFileProperties, alias, secretName, certKey string) {
	// append secret name to the path, so files are not overridden if used same key in the cert
	fileDir := filepath.Join(Path + secretName)
	filePath := filepath.Join(fileDir, certKey)

	dockerFileProp.Certs[alias] = filePath
	dockerFileProp.CertFound = true

	// add volumes
	vol, mount := k8s.SecretVolumeMount(secretName, fileDir, "")
	kaniko.AddVolume(vol, mount)
}
