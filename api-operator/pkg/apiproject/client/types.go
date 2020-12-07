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

package client

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject/build"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
)

// AdapterClient is the interface for clients to update API projects to Adapter
type AdapterClient interface {
	Update(ctx context.Context, reqInfo *common.RequestInfo, projects *build.ProjectsMap) (Response, error)
}

// Response represents the response code list after updating the microgateway
// maps [project -> response code]
//
// a_com
//   Updated
// b_com
//   Failed
// c_com
//   Deleted
//
type Response map[string]ResponseType

// ResponseType represents the response of Failed, Updated, Deleted of Adapter
type ResponseType string

const (
	Failed  = ResponseType("Failed")
	Updated = ResponseType("Updated")
	Deleted = ResponseType("Deleted")
)
