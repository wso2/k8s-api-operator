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

package envoy

import (
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/utils"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logSendAPIs = log.Log.WithName("mgw.envoy.sendAPIs")

func CreateFileToSend(apiList *wso2v1alpha2.APIList, client *client.Client) error {
	var fileName string
	var cleanupFunc func()
	dir, _ := ioutil.TempDir("", "test")

	for _, api := range apiList.Items {
		inputConf := k8s.NewConfMap()
		err := k8s.Get(client, types.NamespacedName{Namespace: api.Namespace,
			Name: api.Spec.SwaggerConfigMapName}, inputConf)
		if err != nil {
			return err
		}
		if inputConf.BinaryData != nil {
			fileName, err = getZipData(inputConf)
			if err != nil {
				return err
			}
		} else {
			fileName, cleanupFunc, err = getSwaggerData(inputConf)
			if err != nil {
				return err
			}
		}
		filePath := filepath.Join(dir, filepath.FromSlash(fileName))
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	tmp, _ := ioutil.TempFile("", "test.zip")
	_ = utils.Zip(dir, tmp.Name())
	testMethod(tmp)
	//cleanup the temporary artifacts once consuming the zip file
	if cleanupFunc != nil {
		defer cleanupFunc()
	}

	return nil
}

// testMethod tests whether the created zip file has data or not (Temporary test method)
//TODO: Send the Zipped APIs file to MGW Adapter
func testMethod(tmp *os.File) {
	filename := tmp.Name()
	data, _ := ioutil.ReadFile(filename)
	if data != nil {
		logSendAPIs.Info("ZIP FILE HAS DATA")
	} else {
		logSendAPIs.Info("NO DATA !!!!")
	}
}
