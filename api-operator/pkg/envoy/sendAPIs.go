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
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logSendAPIs = log.Log.WithName("mgw.envoy.sendAPIs")

// CreateFileToSend creates the zipped file with all the APIs
func CreateFileToSend(apiList *wso2v1alpha2.APIList, client *client.Client) error {
	var fileName string
	var cleanupFunc func()
	var files []string
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
		files = append(files, fileName)
	}

	tmp := os.TempDir() + "/done.zip"
	err := ZipFiles(tmp, files)
	if err != nil {
		logSendAPIs.Error(err, "Error adding the zip files to a single zip file")
		return err
	}
	//TODO: Send APIs set by set when there are many APIs (Eg: 1000s of APIs)
	//// Split the files array into chunks of 100 in size (100 APIs in one chunk)
	//limit := 100
	//for i := 0; i < len(files); i += limit {
	//	batch := files[i:min(i+limit, len(files))]
	//	tmp := fmt.Sprintf(os.TempDir()+ "/done-%d.zip", i)
	//	// Add files to zip
	//	err := ZipFiles(tmp, batch)
	//	if err != nil {
	//		logSendAPIs.Error(err, "Error adding the zip files to a single zip file")
	//		return err
	//	}
	//}

	//cleanup the temporary artifacts once consuming the zip file
	if cleanupFunc != nil {
		defer cleanupFunc()
	}
	return nil
}

//func min(a, b int) int {
//	if a <= b {
//		return a
//	}
//	return b
//}
