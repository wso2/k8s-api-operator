## Scenario 1 - Expose a K8s service as an API

- This scenario describes how to expose a backend service which has been already deployed in the
kubernetes cluster as a managed API in the Kubernetes cluster.
- First we will deploy a sample backend service (product service) in the Kubernetes cluster
- Then the backend service (exposed k8s service) will be exposed as a managed API in the Kubernetes cluster 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the sample backend service
 
 - Navigate to wso2am-k8s-crds-1.0.0/scenarios/scenario-1 directory and deploy the sample backend service using the following command.
 
    ```
        apictl apply -f products_dep.yaml
    ```
    
    - Output:
    ```
        service/products created
        deployment.apps/products-deployment created
    ```
    
 - This will deploy ***products*** backend service on port 80 with the following resources
     - GET ***/products*** : list all the products available
     - GET ***/products/{productId}***   : list product specific details for the given product ID
     
 - Execute the following command to check if the service is present in the Kubernetes cluster.
    ```
        apictl get services products
    ``` 
    
    - Output:
    ```
        NAME       TYPE           CLUSTER-IP    EXTERNAL-IP       PORT(S)        AGE
        products   LoadBalancer   10.83.1.131   104.197.114.248   80:30475/TCP   27m
    ```
    
 - To test if the product service is working, execute the following commands.
    ```
        Command 1:
        curl -X GET http://<EXTERNAL-IP>:80/products
        
        Output:
        {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
    
    ```
        Command 2:
        curl -X GET http://104.197.114.248:80/products/101
        
        Output:
        {"name":"Apples", "id":101, "price":"$1.49 / lb", "reviewScore":"0", "stockAvailability":false}
    ``` 

##### Exposing the backend service as a managed API in the K8S cluster

> Please note that you need to configure the k8s-apim-operator in the Kubernetes cluster prior to trying out the scenarios.

- Basic swagger definition belongs to the "products" service is available in products_swagger.yaml.<br>

    Base path of the API and backend endpoint of the API should be mentioned in the swagger file with the  "x-wso2-basePath" and "x-wso2-production-endpoints" extensions respectively. <br>
    ```$xslt
        x-wso2-basePath: /store/v1.0.0
    ```

    In this swagger definition, the backend service of the "products" service has been mentioned as follows. It can be either globally or resource level.
    In the scenarios, we have defined it in resource level.

    ```
        x-wso2-production-endpoints:
          urls:
            - http://products
    ```

- Create API <br /> 
    ```
        apictl add api -n online-store --from-file=products_swagger.yaml
    ``` 
    - Output:
    ```$xslt
        creating configmap with swagger definition
        configmap/online-store-swagger created
        api.wso2.com/online-store created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
        NAME          AGE
        online-store   55m
    ```

- Get service details to invoke the API<br />
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME                 TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)                         AGE
        online-store        LoadBalancer   10.83.9.188    34.66.153.49      9095:32087/TCP,9090:32572/TCP   98m
        products            LoadBalancer   10.83.1.131    104.197.114.248   80:30475/TCP                    77m
    ```
    - You can see both the backend(products) service and the managed API service(online-store) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjV9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06OTQ0M1wvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoiOFpWV1lQYkk2Rm1lY0ZoeXdVaDVVSXJaNEFvYSIsImV4cCI6MzcxODI5OTU1MiwiaWF0IjoxNTcwODE1OTA1LCJqdGkiOiJkMGI2NTgwNC05NDk3LTQ5ZjktOTcxNC01OTJmODFiNzJhYjMifQ==.HYCPxCbNcALcd0svu47EqFoxnnBAkVJSnCPnW6jJ1lZQTzSAiuiPcGzTnyP1JHodQknhYsSrvdZDIzWzU_mRH2i3-lMVdm0t43r-0Ti0EdBSX2756ilo266MVeWhxbz9p3hPm5ndDCoo_bfB4KbjigjmhXv_PJyUMuWtMo669sHQNs5FkiOT2X0gzFP1iJUFf-H9y762TEIYpylKedVDzQP8x4LCRZsO54e1iA-DZ5h5MKQhJsbKZZ_MMXGmtdo8refPyTCc7HIuevUXIWAaSNRFYj_HZTSRYhFEUtDWn_tJiySn2umRuP3XqxPmQal0SxD7JiV8DQxxyylsGw9k6g==
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/store/v1.0.0/products" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (online-store), which then call its endpoint("products" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
        {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
    

- Delete the  API <br /> 
    ```
        apictl delete api online-store
    ```
    -  Output:
    ```
        api.wso2.com "online-store" deleted
    ```