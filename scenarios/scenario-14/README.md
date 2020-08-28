## Scenario 14 - API Management in Serverless (Knative)
- This scenario describes how to expose a service as a managed API in serverless mode.
- In serverless mode, backend will be deployed in serverless mode as a Knative service.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the serverless mode in the swagger file.
- Later we will deploy the API using the swagger definition. 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to
> "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check
> analytics.

### Installation Prerequisites

- [Kubernetes v1.16 or above](https://Kubernetes.io/docs/setup/) <br>

#### Installing Knative

Follow the document
[Installing the Serving component (version v0.16)](https://knative.dev/v0.16-docs/install/any-kubernetes-cluster/#installing-the-serving-component)
with picking the networking layer as **Istio**. No need to follow the section "Installing the Eventing component" of the
document.

- Monitor the Knative components until all of the components show a **STATUS** of **Running** or **Completed**.
```sh
>> kubectl get pods --namespace knative-serving
```

 #### Deploying the artifacts
 
 **Important:** 
 > If you haven't install the API Operator, please follow the main readme to deploy the API Operator.
 
 - Navigate to `<api-operator-home>/scenarios/scenario-14` directory and deploy the sample backend service in micro namespace.
    ```sh
    >> apictl create ns micro;
       apictl apply -f hello-world-serverless.yaml;
   
    Output:
    namespace/micro created
    targetendpoint.wso2.com/hello-world-serverless created
    ```
   
- Basic swagger definition belongs to the "hello-world-serverless" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
You must mention the TargetEndpoint namespace with the endpoint name as follows  <endpoint-name>.<namespace>
The mode of managed API which is Serverless also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "hello-world-serverless" service and the managed API mode have been mentioned as follows.
    ```
    x-wso2-production-endpoints:
      urls:
        - hello-world-serverless.micro
    x-wso2-mode: Serverless
    ```
  
- Create API <br /> 

    ```sh
    >> apictl create ns api;
       apictl add api -n hello-world \
                    --namespace api \
                    --from-file=swagger.yaml \
                    --override
    
    Output:
    creating configmap with swagger definition
    configmap/hello-world-swagger created
    creating API definition
    api.wso2.com/hello-world created
    ```
  
    **Note:** When you use the --override flag, it builds the docker image and pushes to the docker registry although it is available in the docker registry. If you are using AWS ECR as the registry type, delete the image of the API.
      
    
- Get available API <br /> 
    ```sh
    >> apictl get apis -n api
    
    - Output: 
    NAME          AGE
    hello-world   1m
    ```

- Get service details to invoke the API<br />
    ```sh
    >> apictl get services -n api
 
    - Output:
    NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
    hello-world           LoadBalancer   10.0.30.231     <pending>     9095:30938/TCP,9090:30557/TCP   1s
    ```
    - You can see the managed API service(hello-world) is available.
    - Get the external IP of the managed API's service

- Verify pods in the namespace **micro** before invoking the API.
    ```sh
    >> apictl get pods -n micro
    
    Output:
    No resources found in micro namespace.
    ```
 
- Invoking the API <br />
    ```sh
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
  
    ```sh
    >> curl -X GET "https://<external IP of LB service>:9095/node/1.0.0/hello/node" -H "Authorization:Bearer $TOKEN" -k
    ```
  
    - Once you execute the above command, it will call to the managed API (hello-world), which then call its endpoint("hello-world-serverless" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
    Hello World!
    ```
 
- Knative handles the serverless function according to the request amount taken by the application, If there are zero requests to the application
  Knative will scale down the application to zero, When application starts getting the requests it will scale up your application.    
- List the pods and check how the backend services is running in serverless mode.

    ```sh
    >> apictl get pods -n micro
  
    Output:
    hello-world-serverless-cbjfs-deployment-76447c984c-7wfbd   2/2     Running   0          9s
    ```
- Clean up all resources <br />
    ```sh
    >> apictl delete ns api micro istio-system knative-serving
 
    Output:
    namespace "api" deleted
    namespace "micro" deleted
    namespace "istio-system" deleted
    namespace "knative-serving" deleted
    ```
