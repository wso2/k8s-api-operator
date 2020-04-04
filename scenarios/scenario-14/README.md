## Scenario 14 - API Management in Serverless (Knative)
- This scenario describes how to expose a service as a managed API in serverless mode.
- In serverless mode, backend will be deployed in serverless mode as a Knative service.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the serverless mode in the swagger file.
- Later we will deploy the API using the swagger definition. 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 ##### Installing Istio for Knative
  
  - before installing Knative you need to install Istio for your cluster
  - Using following command you can install Istio for Knative
  - First downlaod and unpack Istio
    ```
    export ISTIO_VERSION=1.3.6
    curl -L https://git.io/getLatestIstio | sh -
    cd istio-${ISTIO_VERSION}
    ```
  - Enter the following command to install the Istio CRDs first
    ```
    >> for i in install/kubernetes/helm/istio-init/files/crd*yaml; do kubectl apply -f $i; done
    ```
  - Create istio-system namespace
    ```
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Namespace
    metadata:
      name: istio-system
      labels:
        istio-injection: disabled
    EOF
    ```
  - Finish Installing Istio with sidecar injection
    ```
    helm template --namespace=istio-system \
      --set sidecarInjectorWebhook.enabled=true \
      --set sidecarInjectorWebhook.enableNamespacesByDefault=true \
      --set global.proxy.autoInject=disabled \
      --set global.disablePolicyChecks=true \
      --set prometheus.enabled=false \
      `# Disable mixer prometheus adapter to remove istio default metrics.` \
      --set mixer.adapters.prometheus.enabled=false \
      `# Disable mixer policy check, since in our template we set no policy.` \
      --set global.disablePolicyChecks=true \
      --set gateways.istio-ingressgateway.autoscaleMin=1 \
      --set gateways.istio-ingressgateway.autoscaleMax=2 \
      --set gateways.istio-ingressgateway.resources.requests.cpu=500m \
      --set gateways.istio-ingressgateway.resources.requests.memory=256Mi \
      `# More pilot replicas for better scale` \
      --set pilot.autoscaleMin=2 \
      `# Set pilot trace sampling to 100%` \
      --set pilot.traceSampling=100 \
      install/kubernetes/helm/istio \
      > ./istio.yaml
    
    >> kubectl apply -f istio.yaml
    ```
  - Use following command to install cluster-local-gateway for Istio
    ```
    helm template --namespace=istio-system \
      --set gateways.custom-gateway.autoscaleMin=1 \
      --set gateways.custom-gateway.autoscaleMax=2 \
      --set gateways.custom-gateway.cpu.targetAverageUtilization=60 \
      --set gateways.custom-gateway.labels.app='cluster-local-gateway' \
      --set gateways.custom-gateway.labels.istio='cluster-local-gateway' \
      --set gateways.custom-gateway.type='ClusterIP' \
      --set gateways.istio-ingressgateway.enabled=false \
      --set gateways.istio-egressgateway.enabled=false \
      --set gateways.istio-ilbgateway.enabled=false \
      --set global.mtls.auto=false \
      install/kubernetes/helm/istio \
      -f install/kubernetes/helm/istio/example-values/values-istio-gateways.yaml \
      | sed -e "s/custom-gateway/cluster-local-gateway/g" -e "s/customgateway/clusterlocalgateway/g" \
      > ./istio-local-gateway.yaml
    
    >> kubectl apply -f istio-local-gateway.yaml
    ```
   
 ##### Installing Knative
 
 - We need to install Knative for deploy our backend service in serverless mode.
 - Using following commands you can install Knative into your cluster.
    ```
     >> kubectl apply --filename https://github.com/knative/serving/releases/download/v0.13.0/serving-crds.yaml
     >> kubectl apply --filename https://github.com/knative/serving/releases/download/v0.13.0/serving-core.yaml
     >> kubectl apply --filename https://github.com/knative/serving/releases/download/v0.13.0/serving-istio.yaml
     >> kubectl apply --filename https://github.com/knative/serving/releases/download/v0.13.0/serving-default-domain.yaml

    ```
   
 - Check all your Knative component is up and running using following command.
    ```` 
     >> kubectl get pods --namespace knative-serving
    ````

 ##### Deploying the artifacts
 
 - Navigate to `<api-operator-home>/scenarios/scenario-14` directory and deploy the sample backend service using the following command.
    ```
     >> apictl apply -f hello-world-serverless.yaml
   
   Output:
       targetendpoint.wso2.com/hello-world-serverless created
    ```
- Basic swagger definition belongs to the "hello-world-serverless" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
You must mention the TargetEndpoint namespace with the endpoint name as follows  <endpoint-name>.<namespace>
The mode of managed API which is Serverless also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "hello-world-serverless" service and the managed API mode have been mentioned as follows.
    ```
        x-wso2-production-endpoints: hello-world-serverless.default
        x-wso2-mode: Serverless
    ```
  
- Create a new namespace for deploy the API

    ```
    >> kubectl create namespace api
    ```
- Create API <br /> 
    ```
    >> apictl add api -n hello-world --from-file=swagger.yaml --namespace=api
    
    - Output:
        creating configmap with swagger definition
        configmap/hello-world-swagger created
        api.wso2.com/hello-world created
    ```
    
- Get available API <br /> 
    ```
    >> apictl get apis --namespace=api
    
    - Output: 
       NAME          AGE
       hello-world   1m
    ```

- Get service details to invoke the API<br />
    ```
    >> apictl get services --namespace=api
 
    - Output:
        NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
        hello-world           LoadBalancer   10.0.30.231     <pending>     9095:30938/TCP,9090:30557/TCP   1s
    ```
    - You can see the managed API service(hello-world) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    >>  curl -X GET "https://<external IP of LB service>:9095/node/1.0.0/hello/node" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (hello-world), which then call its endpoint("hello-world-serverless" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
        Hello World!
    ```
 
- Knative handles the serverless function according to the request amount taken by the application, If there are zero requests to the application
  Knative will scale down the application to zero, When application starts getting the requests it will scale up your application.    
- List the pods and check how the backend services is running in serverless mode.

    ```$xslt
    >> apictl get pods      
  
    - Output:
        hello-world-serverless-cbjfs-deployment-76447c984c-7wfbd   2/2     Running   0          9s
    ```
- Delete the  API <br /> 
    ```
    >>  apictl delete api hello-world --namespace=api
  
    -  Output:
        api.wso2.com "hello-world" deleted
    ```
