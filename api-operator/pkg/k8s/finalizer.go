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

package k8s

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func HandleDeletion(ctx context.Context, requestInfo *common.RequestInfo, finalizer string, handle func(context.Context, *common.RequestInfo) error) (deleted, finalizerUpdated bool, err error) {
	meta := requestInfo.Object.(v1.ObjectMetaAccessor).GetObjectMeta()
	if meta.GetDeletionTimestamp().IsZero() {
		// object is not being deleted
		if !str.ContainsString(meta.GetFinalizers(), finalizer) {
			// add finalizer
			meta.SetFinalizers(append(meta.GetFinalizers(), finalizer))
			if err := requestInfo.Client.Update(ctx, requestInfo.Object); err != nil {
				return false, false, err
			}
			requestInfo.Log.Info("Updated object with adding finalizer", "finalizer", finalizer)
			return false, true, nil
		}
		return false, false, nil
	} else {
		// object is being deleted
		if str.ContainsString(meta.GetFinalizers(), finalizer) {
			// handle finalizer
			requestInfo.Log.V(1).Info("Run finalizer handler before removing the specified finalizer", "finalizer", finalizer, "pending_finalizers", meta.GetFinalizers())
			if err := handle(ctx, requestInfo); err != nil {
				return false, false, err
			}
			// remove finalizer
			meta.SetFinalizers(str.RemoveString(meta.GetFinalizers(), finalizer))
			if err := requestInfo.Client.Update(ctx, requestInfo.Object); err != nil {
				return false, false, err
			}
			requestInfo.Log.Info("Updated object with removing finalizer", "finalizer", finalizer)
			return true, true, nil
		}
		return true, false, nil
	}
}

func IsDeleted(object runtime.Object) bool {
	meta := object.(v1.ObjectMetaAccessor).GetObjectMeta()
	return !meta.GetDeletionTimestamp().IsZero()
}
