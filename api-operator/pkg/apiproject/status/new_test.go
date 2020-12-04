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

package status

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestNewFromIngress(t *testing.T) {
	ing := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ing1",
			Namespace: "foo",
		},
		Spec: v1beta1.IngressSpec{
			IngressClassName: nil,
			Backend:          nil,
			TLS:              nil,
			Rules: []v1beta1.IngressRule{
				{
					Host:             "a.com",
					IngressRuleValue: v1beta1.IngressRuleValue{},
				},
				{
					Host:             "b.com",
					IngressRuleValue: v1beta1.IngressRuleValue{},
				},
			},
		},
		Status: v1beta1.IngressStatus{},
	}
	ingWithAnnotations := &ingress.Ingress{
		Ingress:           *ing,
		ParsedAnnotations: annotations.Ingress{},
	}

	want := &ProjectsStatus{
		"foo/ing1": map[string]string{"ingress-a_com": "_", "ingress-b_com": "_"},
	}

	status := NewFromIngresses(ingWithAnnotations)

	if !reflect.DeepEqual(status, want) {
		t.Errorf("NewFromIngress ingress: %v returned state: %v; want: %v", *ing, *status, *want)
	}
}
