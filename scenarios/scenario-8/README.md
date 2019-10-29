## Scenario 8 - Deploy APIs in k8s in sidecar mode
- This scenario describes how to expose a service as a managed API in sidecar mode.
- In sidecar mode, backend and the managed API will be deployed in the same pod.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the private jet mode in the swagger file.
- Later we will deploy the API using the swagger definition 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to wso2am-k8s-crds-1.0.0/scenarios/scenario-8 directory and deploy the sample backend service using the following command.
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
        inventory-sidecar   ClusterIP      10.99.237.55     <none>        80/TCP                          8m18s
    ```
    - You can see both the managed API service(inventory-sc) and the backend service(inventory-sidecar) are available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
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
- Delete the  API <br /> 
    ```
        apictl delete api inventory-sc
    ```
    -  Output:
    ```
        api.wso2.com "inventory-sc" deleted
    ```
