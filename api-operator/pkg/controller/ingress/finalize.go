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

package ingress

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
)

const (
	// finalizerName represents the name of ingress finalizer handled by this controller
	finalizerName = "wso2.microgateway/ingress.finalizer"
)

func (r *ReconcileIngress) finalizeDeletion(ctx context.Context, requestInfo *common.RequestInfo) error {
	// handle deletion with finalizers to avoid missing ingress configurations deleted while
	// restating controller, or deleted before starting controller.
	//
	// Ingress deletion delta change also handled in the update delta change flow and
	// skipping handling deletion here
	ingresses, err := getSortedIngressList(ctx, requestInfo)
	if err != nil {
		return err
	}

	if err := r.handleRequest(ctx, requestInfo, ingresses); err != nil {
		return nil
	}

	successfullyHandledRequestCount++
	return nil
}
