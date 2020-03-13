# API Management for Istio Services

In this scenario, we have several microservices which are deployed in Istio. For applying API management for those microservices, we can expose an API for those microservices. 

This works in Istio permissive mode and Strict MTLS mode.

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/) 

- Istio

- An account in DockerHub or private docker registry

- Download [api-k8s-crds-1.0.1.zip](https://github.com/wso2/k8s-apim-operator/releases/download/v1.0.1/api-k8s-crds-1.0.1.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract api-k8s-crds-1.0.1.zip
    
    ```
    cd api-k8s-crds-1.0.1
    ```

### Step 1: Install API Operator

* Deploy the Controller artifacts <br>
This will deploy the artifacts related to the API Operator
    ```
    >> kubectl apply -f apim-operator/controller-artifacts/
    
    Output:
    
    namespace/wso2-system created
    deployment.apps/apim-operator created
    clusterrole.rbac.authorization.k8s.io/apim-operator created
    clusterrolebinding.rbac.authorization.k8s.io/apim-operator created
    serviceaccount/apim-operator created
    customresourcedefinition.apiextensions.k8s.io/apis.wso2.com created
    customresourcedefinition.apiextensions.k8s.io/ratelimitings.wso2.com created
    customresourcedefinition.apiextensions.k8s.io/securities.wso2.com created
    customresourcedefinition.apiextensions.k8s.io/targetendpoints.wso2.com created
    ```

* Deploy the controller level configurations **[IMPORTANT]**

    When you deploy an API, this will create a docker image for the API and be pushed to Docker-Hub. For this, your Docker-Hub credentials are required.
    
    1. Open **apim-operator/controller-configs/controller_conf.yaml** and navigate to docker registry section(mentioned below), and  update ***user's docker registry***.
            
        ```
        #docker registry name which the mgw image to be pushed.  eg->  dockerRegistry: username
        dockerRegistry: <username-docker-registry>
        ```
        
    2. Open **apim-operator/controller-configs/docker_secret_template.yaml** and navigate to data section. <br>
        Enter the base 64 encoded username and password of the Docker-Hub account 
        
        ```
        data:
         username: ENTER YOUR BASE64 ENCODED USERNAME
         password: ENTER YOUR BASE64 ENCODED PASSWORD
        ```
        Once you done with the above configurations, execute the following command to deploy controller configurations.

        ```
        >> kubectl apply -f apim-operator/controller-configs/
        
        configmap/controller-config created
        configmap/apim-config created
        security.wso2.com/default-security-jwt created
        secret/wso2am310-secret created
        configmap/docker-secret-mustache created
        secret/docker-secret created
        configmap/dockerfile-template created
        configmap/mgw-conf-mustache created
        ```
<br />

### Step 2: Install Kubectl extensions

- Make the extensions executable using the following command.
    ```
    >> chmod +x apim-operator/kubectl-extensions/kubectl-*
    ```
- Copy the extensions into ***/usr/local/bin*** directory.
    ```
    >> cp apim-operator/kubectl-extensions/kubectl-* /usr/local/bin/
    ```
    - **NOTE**: You may need to execute the COPY command with ***sudo***.
    
### Step 3: Deploy Microservices

- Create the namespace micro and enable Istio injection.

     ```
     >> kubectl create ns micro
     >> kubectl label namespace micro istio-injection=enabled
     ```
- Deploy microservices

The following artifacts resides in the demo repo. <br/>

    >> kubectl create -f microservices.yaml

    
### Step 4: Deploy an API for the microservices

     
    >> kubectl create ns wso2
    >> kubectl add api online-store-api --from-file=swagger.yaml --namespace=wso2 
   
### Step 5: Setup routing in Istio

Due to Strict MTLS in Istio, we are deploying a gateway and a virtual service in Istio.

    >> kubectl create -f gateway.yaml
    >> kubectl create -f virtualservice.yaml
   
### Step 6: Invoke the API
 
 
 - Retrieve the API service endpoint details
 
     The API service is exposed as the Load Balancer service type. You can get the API service endpoint details by using the following command.
 
     ```
     >> kubectl get services -n wso2
     
     Output:
     NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
     online-store-api       LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
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
          
     >> curl -X GET "https://35.232.188.134:9095/store/v1.0.0/review/101" -H "Authorization:Bearer $TOKEN" -k
     
     >> curl -X GET "https://35.232.188.134:9095/store/v1.0.0/inventory/101" -H "Authorization:Bearer $TOKEN" -k
     ```

 <br />
