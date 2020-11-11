package ingress

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	gwclient "github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/client"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/status"
	inghandler "github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/handler"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
	"time"
)

// TODO (renuka) have to update tests with adding k8s service objects

func TestReconcile(t *testing.T) {
	ctx := context.Background()
	k8sObjects := make([]runtime.Object, 0, 16)

	// Read 4 ingresses
	ingresses := make([]v1beta1.Ingress, 4, 4)
	ingObj := make([]runtime.Object, 4, 4)
	for i := range ingresses {
		ingObj[i] = &ingresses[i]
	}
	k8sObjects = append(k8sObjects, ingObj...)
	if err := readResources("test_resources/existing/ingresses.yaml", ingObj...); err != nil {
		t.Fatal("Error reading ingress resources")
	}

	// Read status configmap
	statusCm := &v1.ConfigMap{}
	k8sObjects = append(k8sObjects, statusCm)
	if err := readResources("test_resources/existing/configmaps.yaml", statusCm); err != nil {
		t.Fatal("Error reading configmap resource")
	}

	// Read services
	svc := make([]v1.Service, 4, 4)
	svcObj := make([]runtime.Object, 4, 4)
	for i := range svc {
		svcObj[i] = &svc[i]
	}
	k8sObjects = append(k8sObjects, svcObj...)
	if err := readResources("test_resources/existing/services.yaml", svcObj...); err != nil {
		t.Fatal("Error reading service resources")
	}

	// Read secrets
	sec := make([]v1.Secret, 4, 4)
	secObj := make([]runtime.Object, 4, 4)
	for i := range sec {
		secObj[i] = &sec[i]
	}
	k8sObjects = append(k8sObjects, secObj...)
	if err := readResources("test_resources/existing/secrets.yaml", secObj...); err != nil {
		t.Fatal("Error reading secret resources")
	}

	k8sClient := fake.NewFakeClientWithScheme(scheme.Scheme, k8sObjects...)

	r := &ReconcileIngress{
		client:      k8sClient,
		scheme:      scheme.Scheme,
		evnRecorder: &record.FakeRecorder{},
		ingHandler:  &inghandler.Handler{GatewayClient: gwclient.NewFakeAllSucceeded()},
	}
	var request reconcile.Request

	// 1.  Update whole world
	t.Run("Build whole world", func(t *testing.T) {
		for _, ingress := range ingresses {
			request = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ingress.Namespace, Name: ingress.Name}}
			if _, err := r.Reconcile(request); err != nil {
				t.Error("Error building whole world from initial ingresses")
			}

			// The following is not a required feature, but it can void unnecessary update of gateway
			if r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap != nil {
				t.Error("Only last request should consider to build whole world")
			}
		}
		// Since update ingresses with finalizers will result to requeue the updated ingress
		// process them again
		for i, ingress := range ingresses {
			request = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ingress.Namespace, Name: ingress.Name}}
			if _, err := r.Reconcile(request); err != nil {
				t.Error("Error building whole world from initial ingresses")
			}

			// The following is not a required feature, but it can void unnecessary update of gateway
			if i < len(ingresses)-1 && r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap != nil {
				t.Error("Only last request should consider to build whole world")
			}
		}

		projectMap := r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap
		tp := (*projectMap)["ingress-__bar_com"].Type
		if tp != action.Update {
			t.Errorf("Ing 5 project: ingress-__bar_com, action: %v; want: Update", tp)
		}
	})

	// 2.  Add new ingress: ing5
	t.Run("Delta change: Add new ingress", func(t *testing.T) {
		newIng5 := &v1beta1.Ingress{}
		if err := readResources("test_resources/new/new-ing5.yaml", newIng5); err != nil {
			t.Fatal("Error reading ingress resource")
		}
		request = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: newIng5.Namespace, Name: newIng5.Name}}
		if err := k8sClient.Create(ctx, newIng5); err != nil {
			t.Fatal("Error in k8s client; err: ", err)
		}
		// Reconcile will update finalizers and requeue request
		// So handle another reconcile
		if _, err := r.Reconcile(request); err != nil {
			t.Error("Error building delta update")
		}
		if _, err := r.Reconcile(request); err != nil {
			t.Error("Error building delta update")
		}

		projectMap := r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap
		tp := (*projectMap)["ingress-__bar_com"].Type
		if tp != action.Update {
			t.Errorf("Ing 5 project: ingress-__bar_com, action: %v; want: Update", tp)
		}

		testCurrentStatus(k8sClient, t, true, "default/ing5", "ingress-__bar_com")
	})

	// 3.  Update ingress: ing1
	t.Run("Delta change: Update ingress", func(t *testing.T) {
		updateIng1 := &v1beta1.Ingress{}
		if err := readResources("test_resources/new/update-ing1.yaml", updateIng1); err != nil {
			t.Fatal("Error reading ingress resource")
		}
		request = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: updateIng1.Namespace, Name: updateIng1.Name}}
		if err := k8sClient.Update(ctx, updateIng1); err != nil {
			t.Fatal("Error in k8s client; err: ", err)
		}
		// Reconcile will update finalizers and requeue request
		// So handle another reconcile
		if _, err := r.Reconcile(request); err != nil {
			t.Error("Error building delta update; err: ", err)
		}
		if _, err := r.Reconcile(request); err != nil {
			t.Error("Error building delta update; err: ", err)
		}

		projectMap := r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap
		tp := (*projectMap)["ingress-___default"].Type
		if tp != action.Update {
			t.Errorf("Ing 1 project: ingress-___default, action: %v; want: Update", tp)
		}
		tp = (*projectMap)["ingress-__foo_com"].Type
		if tp != action.Update {
			t.Errorf("Ing 1 project: ingress-__foo_com, action: %v; want: Update", tp)
		}
		tp = (*projectMap)["ingress-prod_foo_com"].Type
		if tp != action.Update {
			t.Errorf("Ing 1 project: ingress-prod_foo_com, action: %v; want: Update", tp)
		}
		tp = (*projectMap)["ingress-deprecated_foo_com"].Type
		if tp != action.Delete {
			t.Errorf("Ing 1 project: ingress-deprecated_foo_com, action: %v; want: Delete", tp)
		}

		testCurrentStatus(k8sClient, t, true, "default/ing1", "ingress-___default")
		testCurrentStatus(k8sClient, t, true, "default/ing1", "ingress-__foo_com")
		testCurrentStatus(k8sClient, t, false, "default/ing1", "ingress-prod_foo_com")
		testCurrentStatus(k8sClient, t, false, "default/ing1", "ingress-deprecated_foo_com")
	})

	// 4.  Delete ingress: ing3
	t.Run("Delta change: Delete ingress", func(t *testing.T) {
		deleteIng3 := &v1beta1.Ingress{}
		nsName := types.NamespacedName{Namespace: "default", Name: "ing3"}
		request = reconcile.Request{NamespacedName: nsName}
		if err := k8sClient.Get(ctx, nsName, deleteIng3); err != nil {
			t.Fatal("Error in k8s client; err: ", err)
		}
		deleteIng3.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		if err := k8sClient.Update(ctx, deleteIng3); err != nil {
			t.Fatal("Error in k8s client; err: ", err)
		}
		if _, err := r.Reconcile(request); err != nil {
			t.Error("Error building delta update")
		}

		projectMap := r.ingHandler.GatewayClient.(*gwclient.Fake).ProjectMap
		tp := (*projectMap)["ingress-__bar_com"].Type
		if tp != action.Update {
			t.Errorf("Ing 1 project: ingress-__bar_com, action: %v; want: Update", tp)
		}
		tp = (*projectMap)["ingress-deprecated_bar_com"].Type
		if tp != action.Delete {
			t.Errorf("Ing 1 project: ingress-deprecated_bar_com, action: %v; want: Delete", tp)
		}

		testCurrentStatus(k8sClient, t, false, "default/ing3", "ingress-__bar_com")
		testCurrentStatus(k8sClient, t, false, "default/ing3", "ingress-deprecated_bar_com")
	})
}

func testCurrentStatus(k8sClient client.Client, t *testing.T, shouldExists bool, ing, project string) {
	st, err := status.FromConfigMap(context.TODO(), &common.RequestInfo{Client: k8sClient})
	if err != nil {
		t.Fatal("Error reading status from configmap")
	}
	if shouldExists {
		if !st.ContainsProject(ing, project) {
			t.Errorf("\"%v: %v\" should exists in the current status", ing, project)
		}
	} else {
		if st.ContainsProject(ing, project) {
			t.Errorf("\"%v: %v\" should not exists in the current status", ing, project)
		}
	}
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
