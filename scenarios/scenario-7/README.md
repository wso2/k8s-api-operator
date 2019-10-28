## Scenario 7 - Deploy APIs in k8s in private jet mode
- This scenario describes how to expose a service as a managed API in private jet mode.
- In private jet mode, backend and the managed API will be deployed in two different pods.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the private jet mode in the swagger file.
- Later we will deploy the API using the swagger definition 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to wso2am-k8s-crds-1.0.0/scenarios/scenario-7 directory and deploy the sample backend service using the following command.
    ```
        apictl apply -f products-privatejet.yaml
    ```
    - Output:
    ```
        targetendpoint.wso2.com/products-privatejet created
    ```
- Basic swagger definition belongs to the "products" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
The mode of managed API (private jet or sidecar) also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "products" service and the managed API mode have been mentioned as follows.
    ```
        x-wso2-production-endpoints: products-privatejet
        x-wso2-mode: privateJet
    ```

- Create API <br /> 
    ```
        apictl add api -n products-pj --from-file=swagger.yaml
    ``` 
    - Output:
    ```$xslt
        creating configmap with swagger definition
        configmap/products-pj-swagger created
        api.wso2.com/products-pj created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
       NAME          AGE
       products-pj   3m
    ```

- Get service details to invoke the API<br />
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME          TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)                         AGE
        products-pj   LoadBalancer   10.83.2.247   35.232.129.64   9095:31818/TCP,9090:32508/TCP   2m
    ```
    - You can see both the backend(products) service and the managed API service(product-pj) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/storepj/v1.0.0/products" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (product-pj), which then call its endpoint("products" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
        {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
    
- List the pods and check how the backend services and the managed API have been deployed

    ```$xslt
        apictl get pods        
    ```
    - Output:
    ```$xslt
        products-pj-699d65df7f-qt2vv           1/1     Running   0          5m12s
        products-privatejet-6777d6f5bc-gqfg4   1/1     Running   0          25m
        products-privatejet-6777d6f5bc-k88sl   1/1     Running   0          25m
    ```
- Delete the  API <br /> 
    ```
        apictl delete api products-pj
    ```
    -  Output:
    ```
        api.wso2.com "products-pj" deleted
    ```
