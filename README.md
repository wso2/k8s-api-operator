# API Operator for Kubernetes

## Introduction

As microservices are increasingly being deployed on Kubernetes, the need to expose these microservices as well documented, easy to consume, managed APIs is becoming important to develop great applications. The API operator for Kubernetes makes APIs a first-class citizen in the Kubernetes ecosystem. Similar to deploying microservices, you can now use this operator to deploy APIs for individual microservices or compose several microservices into individual APIs. With this users will be able to expose their microservice as managed API in Kubernetes environment without any additional work.


![Alt text](docs/images/API-K8s-Operator.png?raw=true "Title")

## Quick Start Guide

In this document, we will walk through on the following.
- Deploy a sample microservice in Kubernetes
- Install API Operator in Kubernetes
- Install the API portal and security token service
- Configure the API controller
- Expose the sample microservice as a managed API
- Invoke the API
- Push the deployed API to the API portal 
- Generate an access token for the API

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/) <br>
Minimum CPU and Memory for the K8s cluster: **2 vCPU, 8GB of Memory**

- An account in DockerHub or private docker registry

- Download [api-k8s-crds-1.0.1.zip](https://github.com/wso2/k8s-apim-operator/releases/download/v1.0.1/api-k8s-crds-1.0.1.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract api-k8s-crds-1.0.1.zip
    
    ```
    cd api-k8s-crds-1.0.1
    ```
 
**_Note:_** You need to run all commands from within the ***api-k8s-crds-1.0.1*** directory.

<br />

#### Step 1: Deploy a sample microservice in Kubernetes


- Let’s deploy a sample microservice in K8s which lists the details of products. This will deploy a pod and service for the sample service.

    ```
    >> kubectl apply -f scenarios/scenario-1/products_dep.yaml
    service/products created
    deployment.apps/products-deployment created
    ```

    The following command will give you the details of the microservice.

    ```
    >> kubectl get services products
  
    Output:
    NAME       TYPE           CLUSTER-IP    EXTERNAL-IP       PORT(S)        AGE
    products   LoadBalancer   10.83.1.131   104.197.114.248   80:30475/TCP   27m
    ```

<details><summary>If you are using Minikube click here</summary>
<p>

**_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.

- On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.

    ```
    >> minikube service <SERVICE_NAME> --url
    >> minikube service products --url
    ```
    
    The IP you receive from above output can be used as the "external-IP" in the following command.

</p>
</details>

<br>

- To test the microservice, execute the following commands.
    ```
    >> curl -X GET http://<EXTERNAL-IP>:80/products
         
    Output:
    {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
   
    ```
    >> curl -X GET http://<EXTERNAL-IP>:80/products/101
         
    Output:
    {"name":"Apples", "id":101, "price":"$1.49 / lb", "reviewScore":"0", "stockAvailability":false}
    ```
<br />

#### Step 2: Configure API Controller

- Download API controller v3.0.0 for your operating system from the [website](https://wso2.com/api-management/tooling/)

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

You can find available operations using the below command.
```
>> apictl --help
```
By default API controller does not support kubectl command. 
Set the API Controller’s mode to Kubernetes to be compatible with kubectl commands

```
>> apictl set --mode k8s 
```
<br />

#### Step 3: Install API Operator

- Execute the following command to install API Operator interactively and configure repository to push the microgateway image.
- Select "Docker Hub" as the repository type.
- Enter repository name of your Docker Hub account (usually it is the username as well).
- Enter username and the password
- Confirm configuration are correct with entering "Y"

```sh
>> apictl install api-operator
Choose repository type:
1: Docker Hub (Or others, quay.io)
2: Amazon ECR
3: GCR
Choose a number: 1: 1
Enter repository name (john or quay.io/mark): : jennifer
Enter username: : jennifer
Enter password:

Repository: jennifer
Username  : jennifer
Confirm configurations: Y: Y
```

Output:
```sh
[Installing OLM]
customresourcedefinition.apiextensions.k8s.io/clusterserviceversions.operators.coreos.com created
...

[Installing API Operator]
subscription.operators.coreos.com/my-api-operator created
[Setting configs]
namespace/wso2-system created
...

[Setting to K8s Mode]
```
        
#### Step 4: Install the API portal and security token service

Kubernetes installation artifacts for API portal and security token service are available in the k8s-artifacts directory.

The following command will deploy API portal & token service under a namespace called “wso2”. 

```
>> kubectl apply -f k8s-artifacts/api-portal/

Output:
namespace "wso2" created
configmap "apim-conf" created
deployment.apps "wso2apim" created
service "wso2apim" created
```
You can check the details of the running server by checking the status of running pods or services in Kubernetes. 

```
>> kubectl get services -n wso2
NAME       TYPE       CLUSTER-IP   EXTERNAL-IP   PORT(S)                                                           AGE
wso2apim   NodePort   10.97.8.86   <none>        30838:32004/TCP,30801:32003/TCP,32321:32002/TCP,32001:32001/TCP   16s

```

**_Note:_** To access the API portal, add host mapping entries to the /etc/hosts file. As we have exposed the API portal service in Node Port type, you can use the IP address of any Kubernetes node.

```
<Any K8s Node IP>  wso2apim
<Any K8s Node IP>  wso2apim-analytics
```

- For Docker for Mac use "localhost" for the K8s node IP
- For Minikube, use minikube ip command to get the K8s node IP
- For GKE
    ```$xslt
    (kubectl get nodes -o jsonpath='{ $.items[*].status.addresses[?(@.type=="ExternalIP")].address }')
    ```
    - This will give the external IPs of the nodes available in the cluster. Pick any IP to include in /etc/hosts file.
  
   **API Portal** - https://wso2apim:32001/devportal 

<br />

#### Step 5: Expose the sample microservice as a managed API

Let’s deploy an API for our microservice.
The Open API definition of the API can be found in the scenario/scenario-1/products-swagger.yaml.

The endpoint of our microservice is referred in the API definition.

- Deploy the API using the following command

    ```
    >> apictl add api -n "api_name" --from-file="location to the Open API definition"
    
    >> apictl add api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml
    
    Output:
    creating configmap with swagger definition
    configmap/online-store-swagger created
    api.wso2.com/online-store created
    ```

    Optional Parameters
    
    ```
    --replicas=3          Number of replicas
    --namespace=wso2      Namespace to deploy the API
    --override            Overwrite the docker image creation for already created docker image
    
    >> apictl add api -n "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" --namespace="desired namespace"
    ```

    **_Note:_** Namespace and replicas are optional parameters. If they are not provided, the default namespace will be used and 1 replica will be created. 

    When you deploy the API, it will first run the Kaniko job. This basically builds the docker image of the API and pushes it to Docker-Hub. 

    Once the Kaniko job is completed, it will deploy the managed API for your microservice.

- Verify the API deployment

    If you list down the pods immediately after the add API command you will only see the pod related to Kaniko job. Once it is completed you will see the deployed API. If you are on Minikube, this might take several minutes.

    ```
    >> kubectl get pods 
    
    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE
    online-store-kaniko-fxvkt              1/1     Running   0          45s
    ```

    If you execute the same command after sometime you will see the managed API has been deployed after the Kaniko job.

    ```
    >> kubectl get pods 
    
    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE
    online-store-6957fc89d6-kn9sp          1/1     Running   0          21s
    ```

    ```
    >> kubectl get services 
    
    Output:
    NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
    ```

You now have a microgateway deployed in Kubernetes that runs your API for the microservice.

<br />

#### Step 6: Invoke the API


- Retrieve the API service endpoint details

    The API service is exposed as the Load Balancer service type. You can get the API service endpoint details by using the following command.

    ```
    >> kubectl get services
    
    Output:
    NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
    ```

<details><summary>If you are using Minikube click here</summary>
<p>

**_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.

- On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.

    ```
    >> minikube service <SERVICE_NAME> --url
    >> minikube service online-store
    ```
    
    The IP you receive from above output can be used as the "external-IP" in the following command.

</p>
</details>

- Invoke the API as a regular microservice

    Let’s observe what happens if you try to invoke the API as a regular microservice.
    ```
    >> curl -X GET "https://<EXTERNAL-IP>:9095/store/v1.0.0/products" -k
    ```
    
    You will get an error as below.
    
    ```
    {"fault":{"code":900902, "message":"Missing Credentials", "description":"Missing Credentials. Make sure your API invocation call has a header: \"Authorization\""}}
    ```
    
    Since the API is secured now, you are experiencing the above error. Hence you need a valid access token to invoke the API.
    
- Invoke the API with an access token

    You can find a sample token below.
    
    ```
   TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
    Copy and paste the above token in the command line. Now you can invoke the API using the cURL command as below.
    
    ```
    Format: 
    
    >> curl -X GET "https://<EXTERNAL-IP>:9095/<API-context>/<API-resource>" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```

    Example commands:
    
    ```
    >> curl -X GET "https://35.232.188.134:9095/store/v1.0.0/products" -H "Authorization:Bearer $TOKEN" -k
    
    >> curl -X GET "https://35.232.188.134:9095/store/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k
    ```
    
**_Note:_** In a production-level scenario, there should be a way to discover the available services and obtain an access token in a secured manner. For this, we need to push this API to API Portal and get an OAuth 2.0 access token


<br />

#### Step 7: Pushing the API to the API Portal


To make the API discoverable for other users and get the access tokens, we need to push the API to the API portal. Then the app developers/subscribers can navigate to the devportal (https://wso2apim:32001/devportal) to perform the following actions.

- Create an application
- Subscribe the API to the application
- Generate a JWT access token 

The following commands will help you to push the API to the API portal in Kubernetes. Commands of the API Controller can be found [here](https://github.com/wso2/product-apim-tooling/blob/v3.0.0-rc/import-export-cli/docs/apictl.md) 


- Add the API portal as an environment to the API controller using the following command.

    ```
    >> apictl add-env -e k8s --registration https://wso2apim:32001/client-registration/v0.16/register --apim https://wso2apim:32003 --token https://wso2apim:32003/token --admin https://wso2apim:32001/api/am/admin/v0.16 --api_list https://wso2apim:32001/api/am/publisher/v1/apis --app_list https://wso2apim:32001/api/am/store/v1/applications
    
    Output:
    Successfully added environment 'k8s'
    ```

- Initialize the API project using API Controller

    ```
    >> apictl init online-store --oas=./scenarios/scenario-1/products_swagger.yaml
    
    Output:
    Initializing a new WSO2 API Manager project in /home/dinusha/k8s-apim-operator/scenarios/scenario-1/online-store
    Project initialized
    Open README file to learn more
    ```

- Import the API to the API portal. **[IMPORTANT]**

    You need to change the API life cycle status to **PUBLISHED** before importing the API. You can edit the api.yaml file located in online-store/Meta-information/ location.
    For testing purpose use ***admin*** as username and password when prompted.
    </br>
    
    ```
    >> apictl import-api -f online-store/ -e k8s -k
    
    Output:
    Successfully imported API
    ```
<br />

#### Step 8: Generate an access token for the API

By default the API is secured with JWT. Hence a valid JWT token is needed to invoke the API. You can obtain a JWT token using the API Controller command as below.

``` 
>> apictl set --token-type JWT

Output: 
Token type set to: JWT
```
Generate access token for the API with the following command.
```
>> apictl get-keys -n online-store -v v1.0.0 -e k8s --provider admin -k

Output:
API name:  OnlineStore & version:  v1.0.0 exists
API  OnlineStore : v1.0.0 subscribed successfully.
Access Token:  eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJkZWZhdWx0LWFwaWN0bC1hcHAiLCJpZCI6MiwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7IlVubGltaXRlZCI6eyJzdG9wT25RdW90YVJlYWNoIjp0cnVlLCJzcGlrZUFycmVzdExpbWl0IjowLCJzcGlrZUFycmVzdFVuaXQiOm51bGx9fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W3sic3Vic2NyaWJlclRlbmFudERvbWFpbiI6ImNhcmJvbi5zdXBlciIsIm5hbWUiOiJPbmxpbmUtU3RvcmUiLCJjb250ZXh0IjoiXC9zdG9yZVwvdjEuMC4wXC92MS4wLjAiLCJwdWJsaXNoZXIiOiJhZG1pbiIsInZlcnNpb24iOiJ2MS4wLjAiLCJzdWJzY3JpcHRpb25UaWVyIjoiVW5saW1pdGVkIn1dLCJjb25zdW1lcktleSI6Im1Hd0lmUWZuZHdZTVZxT25JVW9Rczhqc1B0Y2EiLCJleHAiOjE1NzIyNjAyMjQsImlhdCI6MTU3MjI1NjYyNCwianRpIjoiNTNlYWJkYWEtY2IyZC00MTQ0LWEzYWUtZDNjNTIxMjgwYjM4In0.QU9rt4WBLcIOXzDkdiBpo_SAN_W4jpMlymPSgdhe4mf4FmdepA6hIXa_NXdzWyOST2XcHskWleL-9bhv4GecvDaCcMUwfSKOo_8DuphYhtv0BukpGpyfzK2SZDtABxxtdRUmNDcyXJiC5NU4laXlDGzUruI_LISjkeeCaK4gA93YQC3Nd0xe14uIO940UNsSiUuI5cZkeKlB9k5vKIzjN1-M-SJCvtDkusvdPTgkSHZL29ICsMQl9rTSRm6dL4xq9rcH7osD-o_amgurkm1RvNagzN0buku6y4tuEyisZvRUlNkQ2KRzX6E6VwNKHAFQ7CG95-k-QYvXDGDXYGNisw  
```
**_Note:_** You also have the option to generate an access token by logging into the devportal.

<br />

#### Documentation

You can find the documentation [here](docs/Readme.md).

#### Cleanup

Execute the following commands if you wish to clean up the Kubernetes cluster by removing all the applied artifacts and configurations related to API operator and API portal.

```
>> apictl delete api online-store
>> apictl delete -f k8s-artifacts/api-portal
>> apictl remove-env -e k8s
>> apictl uninstall api-operator
```

When prompted type `Y` when uninstalling API Operator.
  
#### Sample Scenarios

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
1. [Sample 12: Apply distributed rate-limiting to managed API in Kubernetes cluster](scenarios/scenario-12)
1. [Sample 13: K8s API Operator for Istio](scenarios/scenario-13)

#### Troubleshooting Guide

You can refer [troubleshooting guide](docs/Troubleshooting/troubleshooting.md).
