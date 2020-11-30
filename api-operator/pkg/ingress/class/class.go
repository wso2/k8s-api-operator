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

package class

import (
	"k8s.io/api/networking/v1beta1"
)

const (
	// DefaultName defines the default class used in the WSO2 API Microgateway ingress controller
	DefaultName = "microgateway"
)

var (
	// IngressClass sets the runtime ingress class to use
	// An empty string means accept all ingresses without
	// annotation and the ones configured with class microgateway
	IngressClass = DefaultName
)

// IsValid returns true if the given Ingress specify the ingress.class annotation
// or IngressClassName resource for Kubernetes >= v1.18
func IsValid(ing *v1beta1.Ingress) bool {
	// with annotation
	ingress, ok := ing.Annotations[v1beta1.AnnotationIngressClass]
	if ok {
		//// empty annotation and same annotation on ingress
		//if ingress == "" && IngressClass == DefaultName {
		//	return true
		//}
		//
		//return ingress == IngressClass

		return ingress == IngressClass
	}

	//// 2. k8s < v1.18. Check default annotation
	//if !k8s.IsIngressV1Ready {
	//	return IngressClass == DefaultClass
	//}
	//
	//// 3. without annotation and IngressClass. Check default annotation
	//if k8s.IngressClass == nil {
	//	return IngressClass == DefaultClass
	//}
	//
	//// 4. with IngressClass
	//if ing.Spec.IngressClassName != nil {
	//	return k8s.IngressClass.Name == *ing.Spec.IngressClassName
	//}
	return false
}
