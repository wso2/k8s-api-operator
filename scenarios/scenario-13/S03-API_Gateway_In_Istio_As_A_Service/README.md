# API Gateway in Istio as a Service

In this scenario, we have a microservice deployed in Istio. Also we will deploy the API microgateway in Istio as a normal service.

This works in Istio permissive mode and Strict MTLS mode.

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/) 

- Istio

### Step 1: Deploy the product microservice

- Enable Istio injection for default namespace.

```
>> kubectl label namespace default istio-injection=enabled
```

- Deploy microservices

```
>> kubectl create -f products/
```

### Step 2: Deploy the API microgateway in Istio

```
>> kubectl create -f microgw/mg.yaml
```

### Step 3: Expose API microgateway from Istio ingress gateway

```
>> kubectl create -f wso2-gateway.yaml
```

### Step 4: Invoke the API

- Figureout the IP address of the ingress gateway

Use EXTERNAL-IP as the \<ingress_gateway_host> based on the output of the following command.  
```
>> kubectl get svc istio-ingressgateway -n istio-system
```     

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
 
 >> curl -X GET "http://api.wso2.com/store/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k
 ```


Note: In the microgateway, only 1 API is exposed in this sample. Like in the example, you can deploy multiple microservices in Istio. Then you can expose those microservices via the API microgateway.

API Microgateway in Docker - https://docs.wso2.com/display/MG300/Deploying+the+API+Microgateway+in+Docker

