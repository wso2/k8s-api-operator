## Scenario 18 - Expose an API using Openshift Route

- This scenario showes how to expose a service using Openshift Route.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

***Prerequrireties:***
- Openshift Cluster (v3.11 or higher version).
- If you are in a new project you need to setup following security context constraints.

    - Grant access to the privileged SCC:
        ```
        oc adm policy add-scc-to-user <scc_name> <user_name>
        ```
    - Make sure to grant service accounts access to the privileged SCC.
    
        ```
        oc adm policy add-scc-to-user privileged system:serviceaccount:<project-name>:<service-account>
        ```
    - Add this command to enable container images with any user.
    
        ```
        oc adm policy add-scc-to-group anyuid system:authenticated
        ```   
    - Add this command to enable container images that require root.
    
        ```
        oc adm policy add-scc-to-user anyuid system:serviceaccount:<your-project>:<service-account>
        ```  
- Navigate to the api-operator/controller-artifacts directory and set the operatorMode to "Route" in the 
  controler_conf.yaml file.
  
  ```
  operatorMode: "Route"
  ```
- If you have already deployed the operator you have to update operatorMode to "Route" and apply the changes using
  following command.
  ```
  kubectl apply -f api-operator/controller-artifacts/controler_conf.yaml
  ```
  
##### Deploying the artifacts

- Navigate to scenarios/scenario-17 directory and deploy the sample backend service using the following command.
  ```
   apictl apply -f hello-world-service.yaml
  ```
  - Output:
    ```
    targetendpoint.wso2.com/hello-world-service created
    ```
 - Basic swagger definition belongs to the "hello-world-service" service is available in swagger.yaml.
 - Create an API which is refer to the backend service "hello-world-service" using following command.
   ```
   apictl add api -n products --from-file=swagger.yaml
   ```
   - Output:
       ```
       Processing swagger 1: swagger.yaml
       creating configmap with swagger definition
       configmap/hello-world-1-swagger created
       creating API definition
       api.wso2.com/hello-world created
       ```
 - Get available API
   ```
   apictl get apis
   ```   
   - Output:
        ```
          NAME          AGE
          hello-world    3m
        ```
 - Get available Route service
   ```
   kubectl get routes
   ```
   - Output:
        ```
        NAME             HOSTS                                           PATH         SERVICES      PORT   TERMINATION  WILDCARD  
        hello-world      hello-world-uvindu-k8soperator.apps.novalocal   node/1.0.0   hello-world   9090                None
    
        ```
    - You can see that Route service is available for the service exposed by hello-world.
    - Using the "HOSTS" name and "PATH" of the route resource you can invoke the API.
    - Add the API resource to the curl command before you try to invoke the API.
    
 - Invoking the API 
   ```
   curl http://hello-world-uvindu-k8soperator.apps.novalocal/node/1.0.0/hello/node
   ``` 
   - Once you execute the above command, it will call to the managed API (hello-world), which then call its endpoint("hello-world-service") available in the cluster.
     If the request is success, you would be able to see the response as below.
      
     ````
     Hello World!
     ````
 - Delete the API and sample backend service.
   ```
   kubectl delete api hello-world
   kubectl delete targetendpoint hello-world-service
   ```
   - Output:
     ```
     api.wso2.com "hello-world" deleted
     targetendpoint.wso2.com "hello-world-service" deleted
     ```