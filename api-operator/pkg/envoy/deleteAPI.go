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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logDelete = log.Log.WithName("mgw.envoy.delete")

// TODO: Implement the delete API Function
func DeleteAPIFromMgw (client *client.Client, api *wso2v1alpha2.API) error {
	logDelete.Info("DELETED")
	return nil
}
