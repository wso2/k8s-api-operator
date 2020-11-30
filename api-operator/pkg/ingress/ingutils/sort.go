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

package ingutils

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
)

var log = logf.Log.WithName("controller_ingress")

// SortIngressSlice sorts Ingresses using the CreationTimestamp field
func SortIngressSlice(ingresses []*ingress.Ingress) {
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
