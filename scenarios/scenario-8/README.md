## Scenario 8 - Deploy APIs in k8s in sidecar mode
- This scenario describes how to expose a service as a managed API in sidecar mode.
- In sidecar mode, backend and the managed API will be deployed in the same pod.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the `sidecar` mode in the swagger file.
- Later we will deploy the API using the swagger definition 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to scenarios/scenario-8 directory and deploy the sample backend service using the following command.
    ```
        apictl apply -f inventory-sidecar.yaml
    ```
    - Output:
    ```
        targetendpoint.wso2.com/inventory-sidecar created
    ```
- Basic swagger definition belongs to the "inventory" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
The mode of managed API (private jet or sidecar) also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "products" service and the managed API mode have been mentioned as follows.
    ```
        x-wso2-production-endpoints: inventory-sidecar
        x-wso2-mode: sidecar
    ```

- Create API <br /> 
    ```
        apictl add api -n inventory-sc --from-file=swagger.yaml
    ``` 
    - Output:
    ```
        creating configmap with swagger definition
        configmap/inventory-sc-swagger created
        api.wso2.com/inventory-sc created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
       NAME           AGE
       inventory-sc   3m
    ```

- Get service details to invoke the API<br />
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME                TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)                         AGE
        inventory-sc        LoadBalancer   10.109.107.234   <pending>     9095:31565/TCP,9090:31420/TCP   87s
    ```
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/storesc/v1.0.0/inventory/301" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (inventory-sc), which then call its endpoint ("inventory-sc-service" service) available in the same pod. If the request is success, you would be able to see the response as below.
    ```
        {"status":"SUCCESS","data":"true"}
    ```
    
- List the pods and check how the backend services and the managed API have been deployed

    ```$xslt
        apictl get pods        
    ```
    - Output:
    ```$xslt
        NAME                                    READY   STATUS    RESTARTS   AGE
        inventory-sc-664db4fd5f-xf84l           2/2     Running   0          7m14s
    ```
    - To list the containers running inside the pod
    ```$xslt
        apictl get pods POD_NAME_HERE -o jsonpath='{.spec.containers[*].name}'
    ```
    - Replace the POD_NAME_HERE with the corresponding pod name
    - Output;
    ```$xslt
        inventory-sc-service mgwinventory-sc
    ```
    - Here 
        - inventory-sc-service is the backend service
        - mgwinventory-sc is the managed inventory API
- Delete the  API and the Target Endpoint resource
    ```
        apictl delete api inventory-sc
        apictl delete targetendpoints inventory-sidecar
    ```
    -  Output:
    ```
        api.wso2.com "inventory-sc" deleted
        targetendpoint.wso2.com "inventory-sidecar" deleted
    ```
