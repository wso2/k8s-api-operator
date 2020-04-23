package mgw

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	ingressMode   = "Ingress"
	routeMode     = "Route"
	clusterIPMode = "ClusterIP"

	portConst  = "port"
	httpConst  = "http"
	httpsConst = "https"
)

//Creating a LB balancer service to expose mgw
func Service(api *wso2v1alpha1.API, operatorMode string, owner []metav1.OwnerReference) *corev1.Service {
	var serviceType corev1.ServiceType
	serviceType = corev1.ServiceTypeLoadBalancer

	if strings.EqualFold(operatorMode, ingressMode) || strings.EqualFold(operatorMode, clusterIPMode) ||
		strings.EqualFold(operatorMode, routeMode) {
		serviceType = corev1.ServiceTypeClusterIP
	}

	labels := map[string]string{
		"app": api.Name,
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            api.Name,
			Namespace:       api.Namespace,
			Labels:          labels,
			OwnerReferences: owner,
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{{
				Name:       httpsConst + "-" + portConst,
				Port:       Configs.HttpsPort,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpsPort},
			}, {
				Name:       httpConst + "-" + portConst,
				Port:       Configs.HttpPort,
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: Configs.HttpPort},
			}},
			Selector: labels,
		},
	}

	return svc
}
