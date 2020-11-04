package k8s

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HandleDeletion(requestInfo *common.RequestInfo, finalizer string, handle func(*common.RequestInfo) error) (deleted, finalizerUpdated bool, err error) {
	meta := requestInfo.Object.(v1.ObjectMetaAccessor).GetObjectMeta()
	if meta.GetDeletionTimestamp().IsZero() {
		// object is not being deleted
		if !str.ContainsString(meta.GetFinalizers(), finalizer) {
			// add finalizer
			meta.SetFinalizers(append(meta.GetFinalizers(), finalizer))
			if err := (*requestInfo.Client).Update(requestInfo.Ctx, requestInfo.Object); err != nil {
				return false, false, err
			}
			return false, true, nil
		}
		return false, false, nil
	} else {
		// object is being deleted
		if str.ContainsString(meta.GetFinalizers(), finalizer) {
			// handle finalizer
			if err := handle(requestInfo); err != nil {
				return false, false, err
			}
			// remove finalizer
			meta.SetFinalizers(str.RemoveString(meta.GetFinalizers(), finalizer))
			if err := (*requestInfo.Client).Update(requestInfo.Ctx, requestInfo.Object); err != nil {
				return false, false, err
			}
			return true, true, nil
		}
		return true, false, nil
	}
}
