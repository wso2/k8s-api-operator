package ownerref

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("k8s configmap")

// NewArrayFrom returns an array with a new owner reference object of given meta data
func NewArrayFrom(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) *[]metav1.OwnerReference {
	setOwner := true
	return &[]metav1.OwnerReference{
		{
			APIVersion:         typeMeta.APIVersion,
			Kind:               typeMeta.Kind,
			Name:               objectMeta.Name,
			UID:                objectMeta.UID,
			Controller:         &setOwner,
			BlockOwnerDeletion: &setOwner,
		},
	}
}
