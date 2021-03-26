module github.com/wso2/k8s-api-operator/api-operator

go 1.13

require (
	github.com/Azure/go-autorest/autorest v0.9.3-0.20191028180845-3492b2aff503
	github.com/Azure/go-autorest/autorest/adal v0.8.1-0.20191028180845-3492b2aff503
	github.com/Jeffail/gabs v1.4.0
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/aws/aws-sdk-go v1.29.3 // indirect
	github.com/beorn7/perks v1.0.1
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful v2.13.0+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/getkin/kin-openapi v0.2.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/go-openapi/errors v0.19.6
	github.com/go-openapi/loads v0.19.5
	github.com/go-openapi/runtime v0.19.22
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/swag v0.19.9
	github.com/go-openapi/validate v0.19.10
	github.com/go-resty/resty/v2 v2.4.0
	github.com/go-stack/stack v1.8.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.4.0
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1
	github.com/gophercloud/gophercloud v0.6.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.3
	github.com/heroku/docker-registry-client v0.0.0-20181004091502-47ecf50fd8d4 // indirect
	github.com/imdario/mergo v0.3.8
	github.com/jessevdk/go-flags v1.4.0
	github.com/json-iterator/go v1.1.10
	github.com/magiconair/properties v1.8.1
	github.com/mailru/easyjson v0.7.1
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mitchellh/mapstructure v1.3.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/pavel-v-chernykh/keystore-go/v4 v4.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/prometheus/procfs v0.0.8
	github.com/renstrom/dedent v1.0.0
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/pflag v1.0.5
	github.com/wso2/product-apim-tooling/import-export-cli v0.0.0-20210324085958-e9b02fc0da7f
	go.mongodb.org/mongo-driver v1.3.4
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904
	golang.org/x/mod v0.2.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200403190813-44a64ad78b9b
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	gomodules.xyz/jsonpatch/v2 v2.0.1
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20200720192137-962b7ea3a72a
	istio.io/client-go v0.0.0-20200717004237-1af75184beba
	istio.io/gogo-genproto v0.0.0-20190930162913-45029607206a
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.4
	k8s.io/gengo v0.0.0-20200114144118-36b2048a9120
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
