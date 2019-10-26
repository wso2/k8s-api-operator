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
    ```$xslt
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
        NAME           TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)                         AGE
        inventory-sc   LoadBalancer   10.83.2.247   35.232.129.64   9095:31818/TCP,9090:32508/TCP   2m
    ```
    - You can see the managed API service(inventory-sc) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjV9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06OTQ0M1wvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoiOFpWV1lQYkk2Rm1lY0ZoeXdVaDVVSXJaNEFvYSIsImV4cCI6MzcxODI5OTU1MiwiaWF0IjoxNTcwODE1OTA1LCJqdGkiOiJkMGI2NTgwNC05NDk3LTQ5ZjktOTcxNC01OTJmODFiNzJhYjMifQ==.HYCPxCbNcALcd0svu47EqFoxnnBAkVJSnCPnW6jJ1lZQTzSAiuiPcGzTnyP1JHodQknhYsSrvdZDIzWzU_mRH2i3-lMVdm0t43r-0Ti0EdBSX2756ilo266MVeWhxbz9p3hPm5ndDCoo_bfB4KbjigjmhXv_PJyUMuWtMo669sHQNs5FkiOT2X0gzFP1iJUFf-H9y762TEIYpylKedVDzQP8x4LCRZsO54e1iA-DZ5h5MKQhJsbKZZ_MMXGmtdo8refPyTCc7HIuevUXIWAaSNRFYj_HZTSRYhFEUtDWn_tJiySn2umRuP3XqxPmQal0SxD7JiV8DQxxyylsGw9k6g==
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/store/v1.0.0/inventory/301" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
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
