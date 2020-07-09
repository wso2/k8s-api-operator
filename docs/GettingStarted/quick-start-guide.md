# API Operator for Kubernetes

## Introduction

As microservices are increasingly being deployed on Kubernetes, the need to expose these microservices as well documented, easy to consume, managed APIs is becoming important to develop great applications. The API operator for Kubernetes makes APIs a first-class citizen in the Kubernetes ecosystem. Similar to deploying microservices, you can now use this operator to deploy APIs for individual microservices or compose several microservices into individual APIs. With this users will be able to expose their microservice as managed API in Kubernetes environment without any additional work.


![Alt text](../images/K8s-API-Operator.png?raw=true "K8s API Operator")

## Quick Start Guide

In this document, we will walk through on the following.
- Deploy a sample microservice in Kubernetes
- Configure the API controller
- Install API Operator in Kubernetes
- Install the API portal and security token service
- Expose the sample microservice as a managed API
- Invoke the API
- Push the deployed API to the API portal 
- Generate an access token for the API

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/) <br>

    - Minimum CPU : 6vCPU
    - Minimum Memory : 6GB

- An account in DockerHub or private docker registry

- Download [k8s-api-operator-1.2.0-alpha.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.0-alpha/k8s-api-operator-1.2.0-alpha.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract k8s-api-operator-1.2.0-alpha.zip
    
    ```sh
    cd k8s-api-operator-1.2.0-alpha
    ```
 
    **_Note:_** You need to run all commands from within the ***k8s-api-operator-1.2.0-alpha*** directory.

<br />

#### Step 1: Deploy a sample microservice in Kubernetes


- Let’s deploy a sample microservice in K8s which lists the details of products. This will deploy a pod and service for the sample service.

    ```sh
    >> kubectl apply -f scenarios/scenario-1/products_dep.yaml
    service/products created
    deployment.apps/products-deployment created
    ```

    The following command will give you the details of the microservice.

    ```sh
    >> kubectl get services products
  
    Output:
    NAME       TYPE           CLUSTER-IP    EXTERNAL-IP       PORT(S)        AGE
    products   LoadBalancer   10.83.1.131   104.197.114.248   80:30475/TCP   27m
    ```

    <details><summary>If you are using Minikube click here</summary>
    <p>
    
    **_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.
    
    - On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.
    
        ```sh
        >> minikube service <SERVICE_NAME> --url
        >> minikube service products --url
        ```
        
        The IP you receive from above output can be used as the "external-IP" in the following command.
    
    </p>
    </details>
-----

<br>

- To test the microservice, execute the following commands.
    ```sh
    >> curl -X GET http://<EXTERNAL_IP>:80/products
         
    Output:
    {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
   
    ```sh
    >> curl -X GET http://<EXTERNAL_IP>:80/products/101
         
    Output:
    {"name":"Apples", "id":101, "price":"$1.49 / lb", "reviewScore":"0", "stockAvailability":false}
    ```
<br />

#### Step 2: Configure API Controller

- Download API controller v3.1.0 or the latest v3.1.x from the [API Manager Tooling web site](https://wso2.com/api-management/tooling/)
    
    - Under Dev-Ops Tooling section, you can download the tool based on your operating system.

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

- You can find available operations using the below command.
    
  ```sh
  >> apictl --help
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
    Choose registry type:
    1: Docker Hub
    2: Amazon ECR
    3: GCR
    4: HTTP Private Registry
    5: HTTPS Private Registry
    6: Quay.io
    Choose a number: 1: 1
    Enter repository name: docker.io/jennifer
    Enter username: jennifer
    Enter password: *******
    
    Repository: docker.io/jennifer
    Username  : jennifer
    Confirm configurations: Y: Y
    
  
    Output:
    customresourcedefinition.apiextensions.k8s.io/apis.wso2.com created
    customresourcedefinition.apiextensions.k8s.io/ratelimitings.wso2.com created
    ...
    
    namespace/wso2-system created
    deployment.apps/api-operator created
    ...
    
    [Setting to K8s Mode]
    ```
    
<br />

#### Step 4: Install the API portal and security token service

[WSO2AM Kubernetes Operator](https://github.com/wso2/k8s-wso2am-operator) is used to deploy API portal and security token service. 

- Install the WSO2AM Operator in Kubernetes.

    ```sh
    >> apictl install wso2am-operator
    
    namespace/wso2-system created
    serviceaccount/wso2am-pattern-1-svc-account created
    ...
    configmap/wso2am-p1-apim-2-conf created
    configmap/wso2am-p1-mysql-dbscripts created
    [Setting to K8s Mode]
    ```

- Install API Portal and security token service under a namespace called "wso2"

    ```sh
    >> apictl apply -f k8s-artifacts/wso2am-operator/api-portal/
    
    Output:
    namespace/wso2 created
    configmap/apim-conf created
    apimanager.apim.wso2.com/custom-pattern-1 created
    ```

- Access API Portal and security token service

    ```sh
    >> apictl get pods -n wso2
    
    Output:
    NAME                                                        READY   STATUS    RESTARTS   AGE
    all-in-one-api-manager-5694f99754-4zhpq                     1/1     Running   0          2m39s
        
    >> apictl get services -n wso2
    
    Output:
    NAME       TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)                                                       AGE
    wso2apim   NodePort   10.0.39.192   <none>        8280:32004/TCP,8243:32003/TCP,9763:32002/TCP,9443:32001/TCP   114s
    ```

    **_Note:_** To access the API portal, add host mapping entry to the /etc/hosts file. As we have exposed the API portal service in Node Port type, you can use the IP address of any Kubernetes node.
    
    ```sh
    <ANY_K8S_NODE_IP>  wso2apim
    ```
    
    - For Docker for Mac use "127.0.0.1" for the K8s node IP
    - For Minikube, use minikube ip command to get the K8s node IP
    - For GKE
        ```$xslt
        (apictl get nodes -o jsonpath='{ $.items[*].status.addresses[?(@.type=="ExternalIP")].address }')
        ```
        - This will give the external IPs of the nodes available in the cluster. Pick any IP to include in /etc/hosts file.
      
       **API Portal** - https://wso2apim:32001/devportal 

<br />

#### Step 5: Expose the sample microservice as a managed API

Let’s deploy an API for our microservice. The Open API definition of the API can be found in the scenario/scenario-1/products-swagger.yaml.

The endpoint of our microservice is referred in the API definition.

- Deploy the API using the following command

    ```sh
    >> apictl add api -n <API_NAME> --from-file=<LOCATION_TO_THE_OPEN_API_DEFINITION>
    
    >> apictl add api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml
    
    Output:
    creating configmap with swagger definition
    configmap/online-store-swagger created
    api.wso2.com/online-store created
    ```

    Optional Parameters
    
    ```sh
    --replicas=3          Number of replicas
    --namespace=wso2      Namespace to deploy the API
    --override            Overwrite the docker image creation for already created docker image
    
    >> apictl add api -n <API_NAME> --from-file=<LOCATION_TO_THE_OPEN_API_DEFINITION> --replicas=<NUMBER_OF_REPLICAS> --namespace=<DESIRED_NAMESPACE>
    ```

    **_Note:_** Namespace and replicas are optional parameters. If they are not provided, the default namespace will be used and 1 replica will be created. 

    When you deploy the API, it will first run the Kaniko job. This basically builds the docker image of the API and pushes it to Docker-Hub. 

    Once the Kaniko job is completed, it will deploy the managed API for your microservice.

- Verify the API deployment

    If you list down the pods immediately after the add API command you will only see the pod related to Kaniko job. Once it is completed you will see the deployed API. If you are on Minikube, this might take several minutes.

    ```sh
    >> apictl get pods 
    
    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE
    online-store-kaniko-fxvkt              1/1     Running   0          45s
    ```

    If you execute the same command after sometime you will see the managed API has been deployed after the Kaniko job.

    ```sh
    >> apictl get pods 
    
    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE
    online-store-6957fc89d6-kn9sp          1/1     Running   0          21s
    ```

    ```sh
    >> apictl get services 
    
    Output:
    NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
    ```

    You now have a microgateway deployed in Kubernetes that runs your API for the microservice.

<br />

#### Step 6: Invoke the API


- Retrieve the API service endpoint details

    The API service is exposed as the Load Balancer service type. You can get the API service endpoint details by using the following command.

    ```sh
    >> apictl get services
    
    Output:
    NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    online-store       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
    ```

    <details><summary>If you are using Minikube click here</summary>
    <p>
    
    **_Note:_**  By default API operator requires the LoadBalancer service type which is not supported in Minikube by default. Here is how you can enable it on Minikube.
    
    - On Minikube, the LoadBalancer type makes the Service accessible through the minikube service command.
    
        ```sh
        >> minikube service <SERVICE_NAME> --url
        >> minikube service online-store --url
        ```
        
        The IP you receive from above output can be used as the "external-IP" in the following command.
    
    </p>
    </details>

-----

- Invoke the API as a regular microservice

    Let’s observe what happens if you try to invoke the API as a regular microservice.
    ```sh
    >> curl -X GET "https://<EXTERNAL-IP>:9095/store/v1.0.0/products" -k
    ```
    
    You will get an error as below.
    
    ```json
    {"fault":{"code":900902, "message":"Missing Credentials", "description":"Missing Credentials. Make sure your API invocation call has a header: \"Authorization\""}}
    ```
    
    Since the API is secured now, you are experiencing the above error. Hence you need a valid access token to invoke the API.
    
- Invoke the API with an access token

    You can find a sample token below.
    
    ```sh
   TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    ```
    Copy and paste the above token in the command line. Now you can invoke the API using the cURL command as below.
    
    ```sh
    Format: 
    
    >> curl -X GET "https://<EXTERNAL-IP>:9095/<API-context>/<API-resource>"  -H "Authorization:Bearer $TOKEN" -k
    ```

    Example commands:
    
    ```sh
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

    ```sh
    >> apictl add-env -e k8s --apim https://wso2apim:32001 --token https://wso2apim:32001/oauth2/token
    
    Output:
    Successfully added environment 'k8s'
    ```

- Initialize the API project using API Controller

    ```sh
    >> apictl init online-store --oas=./scenarios/scenario-1/products_swagger.yaml --initial-state=PUBLISHED
    
    Output:
    Initializing a new WSO2 API Manager project in /home/wso2/k8s-api-operator/scenarios/scenario-1/online-store
    Project initialized
    Open README file to learn more
    ```

- Import the API to the API portal. **[IMPORTANT]**

    For testing purpose use ***admin*** as username and password when prompted.
    </br>
    
    ```sh
    >> apictl import-api -f online-store/ -e k8s -k
    
    Output:
    Successfully imported API
    ```
<br />

#### Step 8: Generate an access token for the API

- By default the API is secured with JWT. Hence a valid JWT token is needed to invoke the API. You can obtain a JWT token using the API Controller command as below.
    
    ```sh
    >> apictl set --token-type JWT
    
    Output: 
    Token type set to: JWT
    ```
- Generate access token for the API with the following command.

    ```sh
    >> apictl get-keys -n online-store -v v1.0.0 -e k8s --provider admin -k
    
    Output:
    API name:  OnlineStore & version:  v1.0.0 exists
    API  OnlineStore : v1.0.0 subscribed successfully.
    Access Token:  eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJkZWZhdWx0LWFwaWN0bC1hcHAiLCJpZCI6MiwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7IlVubGltaXRlZCI6eyJzdG9wT25RdW90YVJlYWNoIjp0cnVlLCJzcGlrZUFycmVzdExpbWl0IjowLCJzcGlrZUFycmVzdFVuaXQiOm51bGx9fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W3sic3Vic2NyaWJlclRlbmFudERvbWFpbiI6ImNhcmJvbi5zdXBlciIsIm5hbWUiOiJPbmxpbmUtU3RvcmUiLCJjb250ZXh0IjoiXC9zdG9yZVwvdjEuMC4wXC92MS4wLjAiLCJwdWJsaXNoZXIiOiJhZG1pbiIsInZlcnNpb24iOiJ2MS4wLjAiLCJzdWJzY3JpcHRpb25UaWVyIjoiVW5saW1pdGVkIn1dLCJjb25zdW1lcktleSI6Im1Hd0lmUWZuZHdZTVZxT25JVW9Rczhqc1B0Y2EiLCJleHAiOjE1NzIyNjAyMjQsImlhdCI6MTU3MjI1NjYyNCwianRpIjoiNTNlYWJkYWEtY2IyZC00MTQ0LWEzYWUtZDNjNTIxMjgwYjM4In0.QU9rt4WBLcIOXzDkdiBpo_SAN_W4jpMlymPSgdhe4mf4FmdepA6hIXa_NXdzWyOST2XcHskWleL-9bhv4GecvDaCcMUwfSKOo_8DuphYhtv0BukpGpyfzK2SZDtABxxtdRUmNDcyXJiC5NU4laXlDGzUruI_LISjkeeCaK4gA93YQC3Nd0xe14uIO940UNsSiUuI5cZkeKlB9k5vKIzjN1-M-SJCvtDkusvdPTgkSHZL29ICsMQl9rTSRm6dL4xq9rcH7osD-o_amgurkm1RvNagzN0buku6y4tuEyisZvRUlNkQ2KRzX6E6VwNKHAFQ7CG95-k-QYvXDGDXYGNisw  
    ```
  
    **_Note:_** You also have the option to generate an access token by logging into the devportal. 

<br />

### Documentation

You can find the documentation [here](../Readme.md).


### Cleanup

- Execute the following commands if you wish to clean up the Kubernetes cluster by removing all the applied artifacts and configurations related to API operator and API portal.

    ```sh
    >> apictl delete api online-store
    >> apictl delete -f k8s-artifacts/api-portal
    >> apictl remove-env -e k8s
    >> apictl uninstall api-operator
    ```

### Sample Scenarios

1. [Sample 1: Expose a K8s service as an API](../../scenarios/scenario-1)
1. [Sample 2: Deploy pet store service as a managed API in k8s cluster](../../scenarios/scenario-2)
1. [Sample 3: Deploy pet store service as a managed API secured with Basic Auth](../../scenarios/scenario-3)
1. [Sample 4: Deploy pet store service as a managed API secured with JWT](../../scenarios/scenario-4)
1. [Sample 5: Deploy pet store service as a managed API secured with OAuth2](../../scenarios/scenario-5)
1. [Sample 6: Apply rate-limiting to managed API in Kubernetes cluster](../../scenarios/scenario-6)
1. [Sample 7: Deploy APIs in k8s in private jet mode](../../scenarios/scenario-7)
1. [Sample 8: Deploy APIs in k8s in sidecar mode](../../scenarios/scenario-8)
1. [Sample 9: Expose an API with multiple service endpoints](../../scenarios/scenario-9)
1. [Sample 10: Apply interceptors to an API](../../scenarios/scenario-10)
1. [Sample 11: Enabling Analytics for managed API](../../scenarios/scenario-11)
1. [Sample 12: Apply distributed rate-limiting to managed API in Kubernetes cluster](../../scenarios/scenario-12)
1. [Sample 13: K8s API Operator for Istio](../../scenarios/scenario-13)
1. [Sample 14: API Management in Serverless (Knative)](../../scenarios/scenario-14)
1. [Sample 15: Apply Java interceptors to an API](../../scenarios/scenario-15)
1. [Sample 16: Deploy multiple swagger-projects as one API](../../scenarios/scenario-16)
1. [Sample 17: Expose an API using Ingress](../../scenarios/scenario-17)
1. [Sample 18: Expose an API using Openshift Route](../../scenarios/scenario-18)


### Troubleshooting Guide

You can refer [troubleshooting guide](../Troubleshooting/troubleshooting.md).
