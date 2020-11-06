package handler

import (
	"context"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/client"
	"io/ioutil"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
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
	logger := log.Log.WithValues("test_logger", "delta_update")
	var k8sObjects []runtime.Object
	var sortedIngresses []*v1beta1.Ingress

	resource, err := readYamlResource("test_resources/ingresses.yaml")
	if err != nil {
		t.Fatal("Error reading test resource file")
	}

	for _, s := range resource {
		ing := &v1beta1.Ingress{}
		if err := yaml.Unmarshal([]byte(s), ing); err != nil {
			t.Fatal("Error parsing test resource files")
		}
		k8sObjects = append(k8sObjects, ing)
		sortedIngresses = append(sortedIngresses, ing)
	}

	k8sClient := fake.NewFakeClientWithScheme(scheme.Scheme, k8sObjects...)

	reqInfo := &common.RequestInfo{
		Request: reconcile.Request{},
		Ctx:     ctx,
		Client:  &k8sClient,
		Object:  k8sObjects[0],
		Log:     logger,
	}

	ingHandler := Handler{
		GatewayClient: client.NewFakeWithRandomResponse(),
	}

	// should test request to the gateway and final gateway state updated in configmap

	for i := 0; i < 10; i++ {
		_ = ingHandler.UpdateDelta(reqInfo, sortedIngresses)
		projectMap := ingHandler.GatewayClient.(*client.Fake).ProjectMap
		response := ingHandler.GatewayClient.(*client.Fake).Response

		fmt.Println(projectMap)
		fmt.Println(response)
	}

}

func readYamlResource(path string) ([]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := string(bytes)
	return strings.Split(s, "\n---\n"), nil
}
