## Scenario 18 - Expose an API using Openshift Route

- This scenario shows how to expose an API using Openshift Route.

**Important:**
> Follow the main README and deploy the api-operator and configuration files.

**Prerequisites:**

- Openshift Cluster (v3.11 or higher version).

- Grant relevant privileges

    You can skip this section if you are deploying the API Operator via Operators in Openshift console.
      
    - Grant access to the privileged SCC:
        ```
        >> oc adm policy add-scc-to-user <scc_name> <user_name>
        ```
    - Make sure to grant service accounts access to the privileged SCC.
    
        ```
        >> oc adm policy add-scc-to-user privileged system:serviceaccount:<project-name>:<service-account>
        ```
    - Add this command to enable container images with any user.
    
        ```
        >> oc adm policy add-scc-to-group anyuid system:authenticated
        ```   
    - Add this command to enable container images that require root.
    
        ```
        >> oc adm policy add-scc-to-user anyuid system:serviceaccount:<your-project>:<service-account>
        ```  
- Navigate to the api-operator/controller-artifacts directory and set the operatorMode to "Route" in the 
  controler_conf.yaml file.
  
  ```
  operatorMode: "Route"
  ```
- If you have already deployed the operator you have to update operatorMode to "Route" and apply the changes using
  following command.
  ```
  >> apictl apply -f api-operator/controller-artifacts/controler_conf.yaml
  ```
  
#### Deploying the artifacts

- Navigate to scenarios/scenario-17 directory and deploy the sample backend service using the following command.
    
    ```
    >> apictl apply -f hello-world-service.yaml
    
    Output:
    targetendpoint.wso2.com/hello-world-service created
    ```
- Basic swagger definition belongs to the "hello-world-service" service is available in swagger.yaml.
- Create an API which is refer to the backend service "hello-world-service" using following command.
 
    ```
    >> apictl add api -n hello-world-api --from-file=swagger.yaml
    
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/hello-world-1-swagger created
    creating API definition
    api.wso2.com/hello-world created
    ```
   
- List down routes

    ```
    >> apictl get routes
    
    Output:
    NAME                                                HOST/PORT            PATH        SERVICES         PORT         TERMINATION      WILDCARD
    api-operator-route-hello-world-api-node-1.0.0     mgw.route.wso2.com   /node/1.0.0  hello-world-api   9090         edge             None
    ```
      
 - Invoking the API 

    ```
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    ```
    
    ```
    >> curl -H "Host:mgw.route.wso2.com" https://34.67.56.7/node/1.0.0/hello/node -H "Authorization:Bearer $TOKEN" -k
    
    Output:
    Hello World!
    ```

**Notes** 
- Only TLS edge support is provided. 
- Tested in Openshift v4.3.1