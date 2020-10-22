# API Management for Istio Services

In this scenario, we have several microservices which are deployed in Istio. For applying API management for those microservices, we can expose an API for those microservices. 

This works in Istio permissive mode and Strict MTLS mode.

![Alt text](mtls-mode.png?raw=true "Istio in Permissive mode and MTLS mode")

### Installation Prerequisites

- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.15 or above](https://Kubernetes.io/docs/setup/) <br>

    - Minimum CPU : 8vCPU
    - Minimum Memory : 8GB
    
- [Istio v1.6.5 or above](https://istio.io/docs/setup/platform-setup/)

- An account in DockerHub or private docker registry

- Download [k8s-api-operator-1.2.2.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.2/k8s-api-operator-1.2.2.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract k8s-api-operator-1.2.2.zip
    
    ```sh
    cd k8s-api-operator-1.2.2/scenarios/scenario-13/S02-APIM_for_Istio_Services_MTLS
    ```
 
    **_Note:_** You need to run all commands from within the ```S02-APIM_for_Istio_Services_MTLS``` directory.

<br />

#### Step 1: Configure API Controller

- Download API controller v3.2.0 or the latest v3.2.x from the [API Manager Tooling web site](https://wso2.com/api-management/tooling/)

    - Under Dev-Ops Tooling section, you can download the tool based on your operating system.

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

- You can find available operations using the below command.
    
  ```sh
  >> apictl --help
  ```
<br />

#### Step 2: Install API Operator

- Set the operator version as `v1.2.2` by executing following in a terminal.
    ```sh
    >> export WSO2_API_OPERATOR_VERSION=v1.2.2
    ```
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
    Enter repository name: jennifer
    Enter username: jennifer
    Enter password: *******
    
    Repository: jennifer
    Username  : jennifer
    Confirm configurations: Y: Y
    ```

    Output:
    ```
    customresourcedefinition.apiextensions.k8s.io/apis.wso2.com created
    customresourcedefinition.apiextensions.k8s.io/ratelimitings.wso2.com created
    ...
    
    namespace/wso2-system created
    deployment.apps/api-operator created
    ...
    
    [Setting to K8s Mode]
    ```
<br />

#### Step 3: Deploy Microservices

- When you execute this command, it creates a namespace called `micro` and **enable Istio sidecar injection** for that
namespace. Also this deploys 3 microservices.

    ```sh
    >> apictl create -f microservices.yaml
    
    >> apictl get pods -n micro
  
    Output:
    NAME                         READY   STATUS    RESTARTS   AGE
    inventory-7dc5dfdc58-gnxqx   2/2     Running   0          9m
    products-8d478dd48-2kgdk     2/2     Running   0          9m
    review-677dd8fbd8-9ntth      2/2     Running   0          9m
    ```
<br />

#### Step 4: Deploy an API for the microservices

- We are creating a namespace called `wso2` and deploy our API there. In this namespace, we have
**NOT enabled Istio sidecar injection**.

    **Note:** For this sample, using the flag `--override` to update configs, if there are images in the docker registry
    which where created during older versions of API Operator.
   
    ```sh
    >> apictl create ns wso2
    >> apictl add api \
                -n online-store-api-mlts \
                --from-file=./swagger.yaml \
                --namespace=wso2 \
                --override
    
    >> apictl get pods -n wso2
  
    Output:
    NAME                                                             READY   STATUS      RESTARTS   AGE
    online-store-api-mlts-5748695f7b-jxnpf                           1/1     Running     0          14m
    online-store-api-mlts-kaniko-b5hqb                               0/1     Completed   0          14m
    ```
<br />

#### Step 5: Setup routing in Istio

- Due to Strict MTLS in Istio, we are deploying a gateway and a virtual service in Istio.

    ```sh
    >> apictl create -f gateway-virtualservice.yaml
    ```
<br />

#### Step 6: Invoke the API
 
- Retrieve the API service endpoint details
 
     The API service is exposed as the Load Balancer service type. You can get the API service endpoint details by using the following command.
 
     ```sh
     >> apictl get services -n wso2
     
     Output:
     NAME                   TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
     online-store-api-mlts  LoadBalancer   10.83.9.142    35.232.188.134   9095:31055/TCP,9090:32718/TCP   57s
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
 
---
 
 - Invoke the API as a regular microservice
 
     Let’s observe what happens if you try to invoke the API as a regular microservice.
     ```sh
     >> curl -X GET "https://<EXTERNAL-IP>:9095/storemep/v1.0.0/products" -k
     ```
     
     You will get an error as below.
     
     ```json
     {
         "fault": {
             "code": 900902,
             "message": "Missing Credentials",
             "description": "Missing Credentials. Make sure your API invocation call has a header: \"Authorization\""
         }
     }
     ```
     
     Since the API is secured now, you are experiencing the above error. Hence you need a valid access token to invoke the API.
     
 - Invoke the API with an access token
 
     You can find a sample token below.
     
     ```sh
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
     ```
     Copy and paste the above token in the command line. Now you can invoke the API using the cURL command as below.
     
     ```sh
     Format: 
     
     >> curl -X GET "https://<EXTERNAL-IP>:9095/<API-context>/<API-resource>" -H "Authorization:Bearer $TOKEN" -k
     ```
 
     Example commands:
     
     ```sh
     >> curl -X GET "https://35.232.188.134:9095/storemep/v1.0.0/products" -H "Authorization:Bearer $TOKEN" -k
     
     >> curl -X GET "https://35.232.188.134:9095/storemep/v1.0.0/products/101" -H "Authorization:Bearer $TOKEN" -k
          
     >> curl -X GET "https://35.232.188.134:9095/storemep/v1.0.0/review/101" -H "Authorization:Bearer $TOKEN" -k
     
     >> curl -X GET "https://35.232.188.134:9095/storemep/v1.0.0/inventory/101" -H "Authorization:Bearer $TOKEN" -k
     ```

 <br />
