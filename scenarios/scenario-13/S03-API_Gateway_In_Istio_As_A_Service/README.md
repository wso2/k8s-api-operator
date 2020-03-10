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
TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
 ```
 Copy and paste the above token in the command line. Now you can invoke the API using the cURL command as below.
 
 ```
 Format: 
 
 >> curl -X GET "http://api.wso2.com/store/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k
 ```


Note: In the microgateway, only 1 API is exposed in this sample. Like in the example, you can deploy multiple microservices in Istio. Then you can expose those microservices via the API microgateway.

API Microgateway in Docker - https://docs.wso2.com/display/MG300/Deploying+the+API+Microgateway+in+Docker

