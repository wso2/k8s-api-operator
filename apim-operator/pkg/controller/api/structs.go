package api

type XMGWProductionEndpoints struct {
	Urls []string `yaml:"urls"`
}

type ServiceEndpoints struct {
	ServiceName string `yaml:"serviceName"`
}

type PolicyYaml struct {
	ResourcePolicies     []Policy `yaml:"resourcePolicies"`
	ApplicationPolicies  []Policy `yaml:"applicationPolicies"`
	SubscriptionPolicies []Policy `yaml:"subscriptionPolicies"`
}

type Policy struct {
	Count    int    `yaml:"count"`
	UnitTime int    `yaml:"unitTime"`
	TimeUnit string `yaml:"timeUnit"`
}
