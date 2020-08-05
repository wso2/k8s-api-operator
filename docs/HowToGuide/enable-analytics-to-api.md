### Incorporating analytics to the k8s operator

By changing the following artifacts, you can point the API Operator to use the API Analytics which is deployed outside
the Kubernetes cluster or anywhere which is accessible to the API Operator running Kubernetes cluster.

- Create two secrets for the analytics server as follows.

    1. Secret 1: Analytics certificate
    2. Secret 2: Include admin credentials (base64 encoded username and password) and secret name of the secret 1.
    
    Samples can be found in `api-operator/apim-analytics-configs/apim_analytics_secret_template.yaml`
    
- To enable analytics you can change the `apim_analytics_conf.yaml`, `analyticsEnabled` to `true`. Give the name of the secret you created above in the `analyticsSecret` field value.

    Samples can be found `api-operator/apim-analytics-configs/apim_analytics_conf.yaml`

Please refer the ***Scenario 11*** which explained how to enable API analytics for the managed API