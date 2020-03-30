# API Gateway in Istio as a Service

In this scenario, we have a microservice deployed in Istio. Also we will deploy the API microgateway in Istio as a normal service.

This works in Istio permissive mode and Strict MTLS mode.

![Alt text](sidecar-mode.png?raw=true "API microgateway in sidecar mode")

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/) <br>

    - Minimum CPU : 6vCPU
    - Minimum Memory : 8GB
    
- [Istio v1.3.x or above](https://istio.io/docs/setup/platform-setup/)

- An account in DockerHub or private docker registry

- Download [k8s-api-operator-1.1.0-beta.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.1.0-beta/k8s-api-operator-1.1.0-beta.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract k8s-api-operator-1.1.0-beta.zip
    
    ```
    cd k8s-api-operator-1.1.0-beta
    ```
 
**_Note:_** You need to run all commands from within the ***k8s-api-operator-1.1.0-beta*** directory.

<br />

#### Step 1: Configure API Controller

- Download API controller v3.1.0-customized for your operating system from the [github](https://github.com/wso2/K8s-api-operator/tree/v1.1.0-beta)

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

You can find available operations using the below command.
```
>> apictl --help
```

#### Step 2: Install API Operator

Set the environment variable `WSO2_API_OPERATOR_VERSION` with the latest API Operator version.

```sh
>> export WSO2_API_OPERATOR_VERSION=v1.1.0-beta
```

- Execute the following command to install API Operator interactively and configure repository to push the microgateway image.
- Select "Docker Hub" as the repository type.
- Enter repository name of your Docker Hub account (usually it is the username as well).
- Enter username and the password
- Confirm configuration are correct with entering "Y"

```sh
>> apictl install api-operator
Choose registry type:
1: Docker Hub (Or others, quay.io, HTTPS registry)
2: Amazon ECR
3: GCR
4: HTTP Private Registry
Choose a number: 1: 1
Enter repository name (docker.io/john | quay.io/mark | 10.100.5.225:5000/jennifer): docker.io/jennifer
Enter username: jennifer
Enter password: *******

Repository: docker.io/jennifer
Username  : jennifer
Confirm configurations: Y: Y
```

Output:
```sh
customresourcedefinition.apiextensions.k8s.io/apis.wso2.com created
customresourcedefinition.apiextensions.k8s.io/ratelimitings.wso2.com created
...

namespace/wso2-system created
deployment.apps/api-operator created
...

[Setting to K8s Mode]
```

#### Step 3: Deploy the product microservices

    >> apictl create -f microservices.yaml

When you execute this command, it creates a namespace called micro and enable Istio sidecar injection for that namespace. Also this deploys 3 microservices.

    >> apictl get pods -n micro
    Output:
    NAME                         READY   STATUS    RESTARTS   AGE
    inventory-7dc5dfdc58-gnxqx   2/2     Running   0          9m
    products-8d478dd48-2kgdk     2/2     Running   0          9m
    review-677dd8fbd8-9ntth      2/2     Running   0          9m

#### Step 4: Deploy an API for the microservices

    >> apictl add api -n online-store-api --from-file=./swagger.yaml --namespace=micro
    
We are creating a namespace called wso2 and deploy our API there. In this namespace, we have not enabled Istio sidecar injection.
 
    >> apictl get pods -n wso2
    Output:
    NAME                                                        READY   STATUS      RESTARTS   AGE
    online-store-api-5748695f7b-jxnpf                           2/2     Running     0          14m
    online-store-api-kaniko-b5hqb                               0/1     Completed   0          14m

#### Step 5: Setup routing in Istio

Due to Strict MTLS in Istio, we are deploying a gateway and a virtual service in Istio.

    >> apictl create -f gateway-virtualservice.yaml
   
#### Step 6: Invoke the API

- Retrieve the IP address of the Ingress gateway

Use EXTERNAL-IP as the \<ingress_gateway_host> based on the output of the following command.  

    >> apictl get svc istio-ingressgateway -n istio-system
    Output:
    NAME                   TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)                                                                                                                                      AGE
    istio-ingressgateway   LoadBalancer   10.0.32.249   34.67.171.126   15020:30939/TCP,80:30104/TCP,443:31782/TCP,15029:30155/TCP,15030:32662/TCP,15031:31360/TCP,15032:32485/TCP,31400:31905/TCP,15443:32303/TCP   13h
     

- Add /etc/hosts entry with the external IP address as below.

```
EXTERNAL-IP api.wso2.com
``` 

- Invoke the API as a regular microservice
 
You can find a sample token below.
 
 ```
TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
 ```
 Copy and paste the above token in the command line. Now you can invoke the API using the cURL command as below.
 
 ```
 Format: 
 
 >> curl -X GET "http://api.wso2.com/storemep/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k
 >> curl -X GET "http://api.wso2.com/storemep/v1.0.0/inventory/101" -H "Authorization:Bearer $TOKEN" -k
 >> curl -X GET "http://api.wso2.com/storemep/v1.0.0/review/101" -H "Authorization:Bearer $TOKEN" -k
 ```

**Note:** In the microgateway, only 1 API is exposed in this sample. Like in the example, you can deploy multiple microservices in Istio. Then you can expose those microservices via the API microgateway.
