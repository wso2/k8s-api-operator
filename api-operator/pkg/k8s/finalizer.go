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
