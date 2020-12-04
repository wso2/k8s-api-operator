module github.com/wso2/k8s-api-operator/api-operator

go 1.13

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/aws/aws-sdk-go v1.29.3
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful v2.13.0+incompatible // indirect
	github.com/getkin/kin-openapi v0.2.0
	github.com/go-logr/logr v0.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/spec v0.19.4
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/heroku/docker-registry-client v0.0.0-20181004091502-47ecf50fd8d4
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.2.8
	istio.io/api v0.0.0-20200720192137-962b7ea3a72a
	istio.io/client-go v0.0.0-20200717004237-1af75184beba
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.4
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
