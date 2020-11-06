package handler

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/client"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/ingutils"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
)

func TestUpdateDelta(t *testing.T) {
	ctx := context.Background()
	logger := log.Log.WithValues("test_logger", "handle_update")
	ingresses := []*v1beta1.Ingress{{}, {}, {}, {}}
	statusCm := &v1.ConfigMap{}
	k8sObjects := []runtime.Object{ingresses[0], ingresses[1], ingresses[2], ingresses[3], statusCm}
	// Read 4 ingresses
	if err := readResources("test_resources/existing/ingresses.yaml", k8sObjects...); err != nil {
		t.Fatal("Error reading ingress resources")
	}
	// Read status configmap
	if err := readResources("test_resources/existing/configmaps.yaml", statusCm); err != nil {
		t.Fatal("Error reading configmap resource")
	}
	k8sClient := fake.NewFakeClientWithScheme(scheme.Scheme, k8sObjects...)

	// 1.  Add new ingress: ing5
	newIng5 := &v1beta1.Ingress{}
	if err := readResources("test_resources/new/new-ing5.yaml", newIng5); err != nil {
		t.Fatal("Error reading ingress resource")
	}

	ingresses = append(ingresses, newIng5)
	ingutils.SortIngressSlice(ingresses)
	reqInfo := &common.RequestInfo{
		Request: reconcile.Request{NamespacedName: types.NamespacedName{Namespace: newIng5.Namespace, Name: newIng5.Name}},
		Ctx:     ctx,
		Client:  &k8sClient,
		Object:  newIng5,
		Log:     logger,
	}
	_ = k8sClient.Create(ctx, newIng5)
	ingHandler := Handler{
		GatewayClient: client.NewFakeAllSucceeded(),
	}
	_ = ingHandler.UpdateDelta(reqInfo, ingresses)
	projectMap := ingHandler.GatewayClient.(*client.Fake).ProjectMap

	tp := (*projectMap)["ingress-__bar_com"].Type
	if tp != action.Update {
		t.Errorf("Ing 5 project: ingress-__bar_com, action: %v; want: Update", tp)
	}
	// TODO: (renuka) check new status in k8s configmap as well
	// TODO: (renuka) check Open API Spec
}

func readResources(path string, obs ...runtime.Object) error {
	resource, err := readYamlResourceFile(path)
	if err != nil {
		return err
	}

	for i, s := range resource {
		if err := yaml.Unmarshal([]byte(s), obs[i]); err != nil {
			return err
		}
	}
	return nil
}

func readYamlResourceFile(path string) ([]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := string(bytes)
	return strings.Split(s, "\n---\n"), nil
}
