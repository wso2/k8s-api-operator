package mgw

import (
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Responce (client *client.Client, apiInstance *wso2v1alpha1.API, operatorMode string, svc *corev1.Service,
	ingressConfData map[string]string, openshiftConfData map[string]string) string {

	var log = logf.Log.WithName("api.controller")
	var ip string
	if operatorMode == "default" {
		loadBalancerFound := svc.Status.LoadBalancer.Ingress
		ip = ""
		for _, elem := range loadBalancerFound {
			ip += elem.IP
		}
		apiInstance.Spec.ApiEndPoint = ip
		log.Info("IP value is :" + ip)
		log.Info("ENDPOINT value in default mode is ","apiEndpoint",apiInstance.Spec.ApiEndPoint)
	}
	if operatorMode == "ingress" {
		ingressHostConf := ingressConfData[ingressHostName]
		log.Info("Host Name is :" + ingressHostConf)
		apiInstance.Spec.ApiEndPoint = ingressHostConf
		log.Info("ENDPOINT value in ingress mode is","apiEndpoint",apiInstance.Spec.ApiEndPoint)
		ip = "<pending>"
	}
	if operatorMode == "route" {
		routeHostConf := openshiftConfData[routeHost]
		log.Info("Host Name is :" + routeHostConf)
		apiInstance.Spec.ApiEndPoint = routeHostConf
		log.Info("ENDPOINT value in route mode is","apiEndpoint",apiInstance.Spec.ApiEndPoint)
		ip = "<pending>"
	}
	if apiInstance.Spec.ApiEndPoint == "" {
		apiInstance.Spec.ApiEndPoint = "<pending>"
		log.Info("ENDPOINT value after updating is","apiEndpoint" ,apiInstance.Spec.ApiEndPoint)
	}

	return ip
}
