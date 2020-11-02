package k8s

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/str"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HandleDeletion(requestInfo *common.RequestInfo, finalizer string, handle func(*common.RequestInfo) error) (deleted bool, err error) {
	meta := requestInfo.Object.(v1.ObjectMetaAccessor).GetObjectMeta()
	if meta.GetDeletionTimestamp().IsZero() {
		// object is not being deleted
		if !str.ContainsString(meta.GetFinalizers(), finalizer) {
			meta.SetFinalizers(append(meta.GetFinalizers(), finalizer))
			if err := (*requestInfo.Client).Update(requestInfo.Ctx, requestInfo.Object); err != nil {
				return false, err
			}
		}
		return false, nil
	} else {
		// object is being deleted
		if str.ContainsString(meta.GetFinalizers(), finalizer) {
			// handle finalizer
			if err := handle(requestInfo); err != nil {
				return true, err
			}
			// remove finalizer
			meta.SetFinalizers(str.RemoveString(meta.GetFinalizers(), finalizer))
			if err := (*requestInfo.Client).Update(requestInfo.Ctx, requestInfo.Object); err != nil {
				return false, err
			}
		}
		return true, nil
	}
}
