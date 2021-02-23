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

package api

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apim"
	wso2v1alpha2 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha2"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/types"
	log2 "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)
var logFinalize = log2.Log.WithName("Finalizer")

func (r *ReconcileAPI) finalizeDeletion(api *wso2v1alpha2.API) error {

	controlConf := k8s.NewConfMap()
	errConf := k8s.Get(&r.client, types.NamespacedName{Namespace: config.SystemNamespace, Name: controllerConfName},
		controlConf)
	if errConf != nil {
		return errConf
	}

	controlConfigData := controlConf.Data
	// Delete the API from API Manager
	deployAPIMEnabled, err := strconv.ParseBool(controlConfigData[deployAPIMEnabledConst])
	if err != nil {
		logFinalize.Error(err, "Invalid boolean value for deployAPIMEnabled",
			"value", controlConfigData[deployAPIMEnabledConst])
		return  err
	}
	if deployAPIMEnabled {
		deleteErr := apim.DeleteImportedAPI(&r.client, api)
		if deleteErr != nil {
			return  deleteErr
		}
		logFinalize.Info("Successfully deleted the API from APIM", "api_name", api.Name)
	}

	// Delete the API from MGW Adapter
	deployMgwEnabled, err := strconv.ParseBool(controlConfigData[deployAPIToMGWEnabledConst])
	if err != nil {
		logFinalize.Error(err, "Invalid boolean value for deployAPIToMGWEnabled",
			"value", controlConfigData[deployAPIToMGWEnabledConst])
		return err
	}

	if deployMgwEnabled {
		errDeleteAPIFromMgw := envoy.DeleteAPIFromMgw(&r.client, api)
		if errDeleteAPIFromMgw != nil {
			return  errDeleteAPIFromMgw
		}
		logFinalize.Info("Successfully Deleted API from Envoy MGW Adapter", "api_name", api.Name)
	}
	return nil
}
