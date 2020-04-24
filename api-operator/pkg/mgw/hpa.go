package mgw

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

var logHpa = log.Log.WithName("mgw.hpa")

const (
	hpaMaxReplicas                 = "hpaMaxReplicas"
	hpaTargetAverageUtilizationCPU = "hpaTargetAverageUtilizationCPU"
)

type AutoScalarProp struct {
	TargetAverageUtilization *int32
	MinReplicas              *int32
	MaxReplicas              int32
}

func GetHPAProp(api *wso2v1alpha1.API, controlConfigData *map[string]string) (*AutoScalarProp, error) {
	prop := &AutoScalarProp{}

	// MinReplicas
	replicas := int32(api.Spec.Replicas)
	prop.MinReplicas = &replicas

	// TargetAverageUtilization
	val := (*controlConfigData)[hpaTargetAverageUtilizationCPU]
	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		logHpa.Error(err, "Error parsing HPA TargetAverageUtilization", "value", val)
		return nil, err
	}
	avgUtil := int32(intVal)
	prop.TargetAverageUtilization = &avgUtil

	// MaxReplicas
	val = (*controlConfigData)[hpaMaxReplicas]
	intVal, err = strconv.ParseInt(val, 10, 32)
	if err != nil {
		logHpa.Error(err, "Error parsing HPA MaxReplicas", "value", val)
		return nil, err
	}
	prop.MaxReplicas = int32(intVal)

	return prop, nil
}

func HPA(dep *appsv1.Deployment, prop *AutoScalarProp, owner *[]metav1.OwnerReference) *v2beta1.HorizontalPodAutoscaler {
	targetResource := v2beta1.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       dep.Name,
		APIVersion: "apps/v1",
	}
	//CPU utilization
	resourceMetricsForCPU := &v2beta1.ResourceMetricSource{
		Name:                     corev1.ResourceCPU,
		TargetAverageUtilization: prop.TargetAverageUtilization,
	}
	metricsResCPU := v2beta1.MetricSpec{
		Type:     "Resource",
		Resource: resourceMetricsForCPU,
	}
	metricsSet := []v2beta1.MetricSpec{metricsResCPU}

	return &v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dep.Name + "-hpa",
			Namespace:       dep.Namespace,
			OwnerReferences: *owner,
		},
		Spec: v2beta1.HorizontalPodAutoscalerSpec{
			MinReplicas:    prop.MinReplicas,
			MaxReplicas:    prop.MaxReplicas,
			ScaleTargetRef: targetResource,
			Metrics:        metricsSet,
		},
	}
}
