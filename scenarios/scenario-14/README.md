## Scenario 14 - API Management in Serverless (Knative)
- This scenario describes how to expose a service as a managed API in serverless mode.
- In serverless mode, backend will be deployed in serverless mode as a Knative service.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the serverless mode in the swagger file.
- Later we will deploy the API using the swagger definition. 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 ##### Installing Knative
 
 - We need to install Knative for deploy our backend service in serverless mode.
 - Using following commands you can install Knative into your cluster.
    ```
   kubectl apply --selector knative.dev/crd-install=true \
   --filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml \
   --filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml \
   --filename https://github.com/knative/serving/releases/download/v0.8.0/monitoring.yaml
    ```
 - “knative.dev/crd-install=true” flag will prevent race conditions during the install, which
    causes intermittent errors.
 - Enter following commands to finish the installation.
 
    ````
   kubectl apply --filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml \
   --filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml \
   --filename https://github.com/knative/serving/releases/download/v0.8.0/monitoring.yaml
   
 - Check all your Knative components are up and running using following command.
 
    ```` 
   kubectl get pods --namespace knative-serving
   kubectl get pods --namespace knative-eventing
   kubectl get pods --namespace knative-monitoring
    ````
 - We have to inject Istio components to our namespace. Knative is using istio to route traffic to your application. use the
   following command to inject Istio to your default namespace.
   ```
   kubectl label namespace default istio-injectio=enabled
    ```   
 ##### Deploying the artifacts
 
 - Navigate to api-k8s-crds-1.0.1/scenarios/scenario-13 directory and deploy the sample backend service using the following command.
    ```
        apictl apply -f hello-world-serverless.yaml
    ```
    - Output:
    ```
        targetendpoint.wso2.com/hello-world-serverless created
    ```
- Basic swagger definition belongs to the "hello-world-serverless" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
You must mention the TargetEndpoint namespace with the endpoint name as follows  <endpoint-name>.<namespace>
The mode of managed API which is private jet also has to be mentioned in the swagger with the "x-wso2-mode" extension.
To enble the serverless feature you need to set true in the serverless property.
In this swagger definition, the backend service of the "products" service and the managed API mode have been mentioned as follows.
    ```
        x-wso2-production-endpoints: hello-world-serverless.default
        x-wso2-mode: privateJet
        x-wso2-serverless: true
    ```
  
- Create a new namespace for deploy the API

    ```
    kubectl create namespace api
    ```
- Create API <br /> 
    ```
        apictl add api -n hello-world --from-file=swagger.yaml --namespace=api
    ``` 
    - Output:
    ```$xslt
        creating configmap with swagger definition
        configmap/hello-world-swagger created
        api.wso2.com/hello-world created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis --namespace=api
    ```
    - Output:
    ```    
       NAME          AGE
       hello-world   1m
    ```

- Get service details to invoke the API<br />
    ```
        apictl get services --namespace=api
    ```
    - Output:
    
    ```
        NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
        hello-world           LoadBalancer   10.0.30.231     <pending>     9095:30938/TCP,9090:30557/TCP   1s
    ```
    - You can see the managed API service(hello-world) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        curl -X GET "https://<external IP of LB service>:9090/node/1.0.0/hello/node"
    ```
    - Once you execute the above command, it will call to the managed API (hello-world), which then call its endpoint("hello-world-serverless" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
        Hello World!
    ```
 
- Knative handles the serverless function according to the request amount taken by the application, If there are zero requests to the application
  Knative will scale down the application to zero, When application starts getting the requests it will scale up your application.    
- List the pods and check how the backend services is running in serverless mode.

    ```$xslt
        apictl get pods      
    ```
    - Output:
    ```$xslt
        hello-world-serverless-cbjfs-deployment-76447c984c-7wfbd   3/3     Running   0          9s
    ```
- Delete the  API <br /> 
    ```
        apictl delete api hello-world
    ```
    -  Output:
    ```
        api.wso2.com "hello-world" deleted
    ```
