# API Operator for Kubernetes

## Introduction

As microservices are increasingly being deployed on Kubernetes the need to expose these microservices as well documented, easy to consume, managed APIs is becoming important to develop great applications. The API operator for Kubernetes makes APIs a first class citizen in the Kubernetes eco-system. Similar to deploying microservices, you can now use this operator to deploy APIs for individual microservices or compose several microservices into individual APIs. With this users will be able to expose their micro service as managed API in Kubernetes environment without any additional work. 


## Quick Start Guide

In this document we will walk through on the following.
- Deploy sample microservice in Kubernetes
- Install API Operator in k8s
- Install API portal and security token service
- Deploy a sample microservice
- Expose the sample microservice as a managed API
- Push the deployed API to the API portal 
- Get an OAuth2.0 access token for the API
- Invoke the API with the access token

##### Install Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/)
- An account in DockerHub or private docker registry

Minimum CPU and Memory for the K8s cluster : **2 vCPU, 8GB of Memory**

Download [api-k8s-crds-1.0.0.zip](https://github.com/wso2/k8s-apim-operator/releases/download/1.0.0/api-k8s-crds-1.0.0.zip) and extract the zip

1. This zip contains the artifacts that required to deploy in Kubernetes.
2. Extract api-k8s-crds-1.0.0.zip .
```
cd api-k8s-crds-1.0.0
```
 
**_Note:_** You need to run all commands from within the ***api-k8s-crds-1.0.0*** directory.

##### Step 1: Deploy a sample microservice in K8s


Let’s deploy a sample microservice in K8s which lists the details of products. This will deploy a pod and service for the sample service.


```
kubectl apply -f  ./scenarios/scenario-1/products_dep.yaml
```

Following command will give you the details of the microservice.

```
      kubectl get services products
```
 
 - Output:
 ```
      NAME       TYPE           CLUSTER-IP    EXTERNAL-IP       PORT(S)        AGE
      products   LoadBalancer   10.83.1.131   104.197.114.248   80:30475/TCP   27m
  ```
 
> If you are using Minikube
**_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.

On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.
```
minikube service <SERVICE_NAME> --url
minikube service products --url
```


- To test if the microservice, execute the following commands.
 ```


      
      curl -X GET http://<EXTERNAL-IP>:80/products
     
      Output:
      {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
  ```
 
  ```
      
      curl -X GET http://<EXTERNAL-IP>:80/products/101
     
      Output:
      {"name":"Apples", "id":101, "price":"$1.49 / lb", "reviewScore":"0", "stockAvailability":false}
 
```


##### Step 2: Install API Operator

- Deploying Controller artifacts

This will deploy the artifacts related to the API Operator
```
kubectl apply -f apim-operator/controller-artifacts/
```

- Deploying controller level configurations **[IMPORTANT]**

When you deploy an API, this will create a docker image for the API and be pushed to Docker-Hub. For this,  your Docker-Hub credentials are required.

- Open **apim-operator/controller-configs/controller_conf.yaml** and navigate to docker registry section(mentioned below), and  update ***user's docker registry***.
```
#docker registry name which the mgw image to be pushed.  eg->  dockerRegistry: username
dockerRegistry: <username-docker-registry>

```

- Open **apim-operator/controller-configs/docker_secret_template.yaml** and navigate to data section.
Enter the base 64 encoded username and password of the Docker-Hub account 

```
data:
 username: ENTER YOUR BASE64 ENCODED USERNAME
 password: ENTER YOUR BASE64 ENCODED PASSWORD

```
```
kubectl apply -f apim-operator/controller-configs/
```
##### Step 3: Install API portal and security token service

Kubernetes installation artifacts for API portal and security token service are available in k8s-artifacts directory.

Following command will deploy API portal & token service under a namespace called “wso2”. 

Please see below command and its output.

```$xslt
>> kubectl apply -f k8s-artifacts/api-portal/
namespace "wso2" created
configmap "apim-conf" created
deployment.apps "wso2apim" created
service "wso2apim" created
```
You can check the details of the running server by checking the status of running pods or services in Kubernetes. 

```$xslt
>> kubectl get services -n wso2
NAME               TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)                                                       AGE
wso2apim   LoadBalancer   10.83.7.193   34.68.222.92   8280:31491/TCP,8243:30517/TCP,9763:32040/TCP,9443:31599/TCP   2d

```

**Note:** To access the API portal, add host mapping entries to the /etc/hosts file. As we have exposed the api portal service in Node Port type, you can use the IP address of any Kubernetes node.

```
<Any K8s Node IP>  wso2apim
<Any K8s Node IP>  wso2apim-analytics
```

- For Docker for Mac use "localhost" for the K8s node IP
- For Minikube, use minikube ip command to get the K8s node IP
  
   **API Portal** - https://wso2apim:32001/devportal 

##### Step 4: Configure API Controller

- Download API controller v3.0.0 for your operating system from - https://github.com/wso2/product-apim-tooling/releases/

- Extract the API controller distribution and navigate inside extracted folder using command line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

You can find available operations using the below command..
```
apictl --help
```
By default API controller does not support kubectl command. 
Set the API Controller’s mode to Kubernetes to be compatible with kubectl commands

```$xslt
apictl set --mode k8s
```

##### Step 5: Deploy an API in K8s cluster via CRDs

Let’s now deploy an API for our microservice.
Definition of the API can be found in the scenario/scenario-1/products-swagger.yaml. 
Endpoint of our microservice has been referred in the API definition.

- Deploy the API using the following command
```
apictl add api -n "api_name" --from-file="location to the api swagger definition"

>> apictl add api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml
creating configmap with swagger definition
configmap/online-store-swagger created
api.wso2.com/online-store created
```

Optional Parameters

```
--replicas=3          Number of replicas
--namespace=wso2      Namespace to deploy the API
--overwrite=true	  Overwrite the docker image creation for already created docker image

apictl add api -n "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" --namespace="desired namespace"
```

**Note:** Namespace and replicas are optional parameters. If they are not provided, the default namespace will be used and 1 replica will be created. 

When you deploy the API, it will first run the Kaniko job. This basically builds the docker image of the API and pushes it to Docker-Hub. 

Once the Kaniko job is completed, it will deploy the managed API for your microservice.


```$xslt
>> kubectl get pods 

```

If you execute above command immediately after the add API command you will only see the pod related to Kaniko job. Once it is completed you will see the deployed API. 
```
NAME                                   READY   STATUS    RESTARTS   AGE
online-store-kaniko-fxvkt              1/1     Running   0          45s

```

If you are on Minikube, this might take several minutes.
If you execute the same command after sometime you will see the managed API has been deployed after the Kaniko job.

```$xslt
>> kubectl get pods 

```

```
``$xsli
NAME                                   READY   STATUS    RESTARTS   AGE
online-store-6957fc89d6-kn9sp          1/1     Running   0          21s

```

```$xslt
>> kubectl get services 
NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s

You now have a microgateway deployed in Kubernetes that runs your API for the microservice.
```

###### Step 6: Invoke the API



The API service is exposed as the Load Balancer service type. You can get the service endpoint details by using the following command.

> If you are using Minikube
**_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.

On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.
```
minikube service <SERVICE_NAME>
```

```
kubectl get services
```

Sample Output:

```
NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
```

Let’s observe what happens if you try to invoke the API as a regular microservice.
```
 curl -X GET "https://35.232.188.134:9095/store/v1.0.0/products" -k
```
You will get an error as below.

```
{"fault":{"code":900902, "message":"Missing Credentials", "description":"Missing Credentials. Make sure your API invocation call has a header: \"Authorization\""}}
```
Since the API is secured now, you are experiencing the above error. Hence you need a valid access token to invoke the API.

You can find a sample token below.
```
TOKEN=eeyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjE1NzIyNTQ5NjEsImlhdCI6MTU3MjI1MTM2MSwianRpIjoiZGM5Mjk2ZDEtY2U4OC00YjkwLWE4MDEtOGMzOTA4MDU1MjIyIn0.NnyNDS61QEUy1PzU67VV_xz2heJ4whT5QTmznDYR6PvKgA222nQc77kgqFjxqo_1_HrPfS0QSDgTMqNvsw72UBRioCDyNjxGAF7nZAgZ0bWrpBJkENFRL0l104vtkcSn4xcxdTBwvAUkB-AMTqdEy-JCLPRca2nQ-_oTOKsu-Zy89sKI2xvZfqLiAda6vbH4_uIn01vogUH7sSAtUQbiHQpfYPQDpq00ZpynPH3Kr80jCkSyQ92hEIryoC_uhsRIKcSS6N6Scc9N-lak7jlxQNs0agqzzIO9cZVzSfnxIANfnFmL5FPacSEJNaZjkU1o8uKwbgob09j_qq-5R9O4Hw

```
Copy and paste the above token in the command line.
Now you can now invoke the API using cURL as below

Format
```
curl -X GET "<EXTERNAL-IP>:<MICROGATEWAY-PORT>/<API-context>/<API-resource>" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
```

Examples

```
curl -X GET "https://35.232.188.134:9095/store/v1.0.0/products" -H "Authorization:Bearer $TOKEN" -k

curl -X GET "https://35.232.188.134:9095/store/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k

```

In a production level scenario, there should be a way to discover the available services and obtain an access token in a secure manner.
For this, we need to push this API to API Portal and get an OAuth2.0 access token
Please refer to the step 7 for pushing API to API portal.

##### Step 7: Pushing the API to the API Portal


To make the API discoverable for other users and get the access tokens, we need to push the API to API portal. Then the app developers/subscribers can navigate to the devportal (https://wso2apim:9443/devportal),
Create an application
Subscribe the API to the application
Generate a JWT access token 

<br>Following commands will help you to push the API to API portal in Kubernetes.

Commands of the API Controller can be found [here](https://github.com/wso2/product-apim-tooling/blob/v3.0.0-beta/import-export-cli/docs/apictl.md) 



First, let’s add the API portal to the API controller using the following command.


```
apictl add-env -e k8s --registration https://wso2apim:32001/client-registration/v0.15/register --apim https://wso2apim:32003 --token https://wso2apim:32003/token --admin https://wso2apim:32001/api/am/admin/v0.15 --api_list https://wso2apim:32001/api/am/publisher/v0.15/apis --app_list https://wso2apim:32001/api/am/store/v0.15/applications

```
Output:
```
Successfully added environment 'k8s'
```

Init the API project using API Controller

```
apictl init online-store --oas=./scenarios/scenario-1/products_swagger.yaml
```
```
Initializing a new WSO2 API Manager project in /home/dinusha/wso2am-k8s-crds-1.0.0/scenarios/scenario-1/online-store
Project initialized
Open README file to learn more
```

Import the API to the API portal. **[IMPORTANT]**

For testing purpose, use *admin* as the username and password of the API portal when prompted. 

**You need to change the API life cycle status to ***PUBLISHED*** before importing the API. You can edit the api.yaml file located in online-store/Meta-information/**
```
apictl import-api -f online-store/ -e k8s -k

```
```
Successfully imported API
```

###### Step 8: Generate an access token for the API

By default the API is secured with JWT. 
Hence a valid JWT token is needed to invoke the API.
You can obtain a JWT token using the API Controller command as below.

``` 
   apictl set --token-type JWT
```
```
  Token type set to: JWT
```

```
apictl get-keys -n online-store -v v1.0.0 -e k8s --provider admin -k
```
```
API name:  OnlineStore & version:  v1.0.0 exists
API  OnlineStore : v1.0.0 subscribed successfully.
Access Token:  eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjE1NzIyNTQ5NjEsImlhdCI6MTU3MjI1MTM2MSwianRpIjoiZGM5Mjk2ZDEtY2U4OC00YjkwLWE4MDEtOGMzOTA4MDU1MjIyIn0.NnyNDS61QEUy1PzU67VV_xz2heJ4whT5QTmznDYR6PvKgA222nQc77kgqFjxqo_1_HrPfS0QSDgTMqNvsw72UBRioCDyNjxGAF7nZAgZ0bWrpBJkENFRL0l104vtkcSn4xcxdTBwvAUkB-AMTqdEy-JCLPRca2nQ-_oTOKsu-Zy89sKI2xvZfqLiAda6vbH4_uIn01vogUH7sSAtUQbiHQpfYPQDpq00ZpynPH3Kr80jCkSyQ92hEIryoC_uhsRIKcSS6N6Scc9N-lak7jlxQNs0agqzzIO9cZVzSfnxIANfnFmL5FPacSEJNaZjkU1o8uKwbgob09j_qq-5R9O4Hw  

```
**Note:** You also have the option to generate an access token by logging into the devportal.

##### Deleting API

Following command will delete the deployed API and artifacts related to that API (pods, deployment, service, HPA)
```$xslt
apictl delete api online-store
```
Output:
```
api.wso2.com "online-store" deleted
```

##### Cleanup

Execute the following command if you want to clean up the Kubernetes cluster by removing all the applied artifacts and configurations related to API operator and API portal.
```
kubect delete -f api-analytics/k8s-artifacts/
kubect delete -f api-manager/k8s-artifacts/
apictl delete -f ./deploy/controller-configs/
apictl delete -f ./deploy/controller-artifacts/
```
  
##### Sample Scenarios

1. [Sample 1: Expose a K8s service as an API](scenarios/scenario-1)
1. [Sample 2: Deploy pet store service as a managed API in k8s cluster](scenarios/scenario-2)
1. [Sample 3: Deploy pet store service as a managed API secured with Basic Auth](scenarios/scenario-3)
1. [Sample 4: Deploy pet store service as a managed API secured with JWT](scenarios/scenario-4)
1. [Sample 5: Deploy pet store service as a managed API secured with OAuth2](scenarios/scenario-5)
1. [Sample 6: Apply rate-limiting to managed API in Kubernetes cluster](scenarios/scenario-6)
1. [Sample 7: Deploy APIs in k8s in private jet mode](scenarios/scenario-7)
1. [Sample 8: Deploy APIs in k8s in sidecar mode](scenarios/scenario-8)
1. [Sample 9: Expose an API with multiple service endpoints](scenarios/scenario-9)
1. [Sample 10: Apply interceptors to an API](scenarios/scenario-10)
1. [Sample 11: Enabling Analytics for managed API](scenarios/scenario-11)


##### Troubleshooting Guide

You can refer [troubleshooting guide](docs/Troubleshooting/troubleshooting.md).








