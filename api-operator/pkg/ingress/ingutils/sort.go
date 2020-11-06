package ingutils

import (
	"fmt"
	"k8s.io/api/networking/v1beta1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
)

var log = logf.Log.WithName("controller_ingress")

// SortIngressSlice sorts Ingresses using the CreationTimestamp field
func SortIngressSlice(ingresses []*v1beta1.Ingress) {
	sort.SliceStable(ingresses, func(i, j int) bool {
		it := ingresses[i].CreationTimestamp
		jt := ingresses[j].CreationTimestamp
		if it.Equal(&jt) {
			in := fmt.Sprintf("%v/%v", ingresses[i].Namespace, ingresses[i].Name)
			jn := fmt.Sprintf("%v/%v", ingresses[j].Namespace, ingresses[j].Name)
			log.V(3).Info("Ingresses have identical CreationTimestamp", "ingress_1", in, "ingress_2", jn)
			return in > jn
		}
		return it.Before(&jt)
	})
}
