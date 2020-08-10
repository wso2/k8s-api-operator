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
    >> apictl get po -n micro
    
    Output:
    No resources found in micro namespace.
    ```
 
- Invoking the API <br />
    ```sh
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
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
- Delete the API <br /> 
    ```sh
    >> apictl delete api hello-world
 
    Output:
    api.wso2.com "hello-world" deleted
    ```
