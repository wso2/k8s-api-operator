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

package security

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logCred = log.Log.WithName("security.credentials")

func SetCredentials(client *client.Client, securityType string, namespacedName types.NamespacedName, mgConfigs *mgw.Configuration) error {
	sha1Hash := sha1.New()
	var userName string
	var password []byte

	//get the secret included credentials
	credentialSecret := k8s.NewSecret()
	err := k8s.Get(client, namespacedName, credentialSecret)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	//get the username and the password
	for k, v := range credentialSecret.Data {
		if strings.EqualFold(k, "username") {
			userName = string(v)
		}
		if strings.EqualFold(k, "password") {
			password = v
		}

	}
	if securityType == "Basic" {

		mgConfigs.BasicUsername = userName
		_, err := sha1Hash.Write(password)
		if err != nil {
			logCred.Info("error in encoding password")
			return err
		}
		//convert encoded password to a hex string
		mgConfigs.BasicPassword = hex.EncodeToString(sha1Hash.Sum(nil))
	}
	if securityType == "Oauth" {
		mgConfigs.KeyManagerUsername = userName
		mgConfigs.KeyManagerPassword = string(password)
	}
	return nil
}
