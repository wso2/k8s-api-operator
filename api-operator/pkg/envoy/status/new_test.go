package status

import (
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestNewFromIngress(t *testing.T) {
	ingress := &v1beta1.Ingress{
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

	want := &ProjectsStatus{
		"foo/ing1": map[string]string{"ingress-a_com": "_", "ingress-b_com": "_"},
	}

	status := NewFromIngress(ingress)

	if !reflect.DeepEqual(status, want) {
		t.Errorf("NewFromIngress ingress: %v returned state: %v; want: %v", *ingress, *status, *want)
	}
}
