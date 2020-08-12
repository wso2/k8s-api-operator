## Scenario 9 - Multiple microservices endpoints as a single managed API
- This scenario describes how to expose a multiple endpoint services via a single managed API.
- First, we will deploy targetendpoint resources for different backend services(products, inventory & review service) on the Kubernetes cluster
- Created target endpoints will be referred accordingly in the swagger definition.
- Then the finalized swagger definition will be deployed a managed API on Kubernetes cluster.

> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

  ##### Deploying the artifacts

 - Navigate to scenarios/scenario-9 directory
 - Deploy the backend services for products, review & inventory
    ```
    apictl apply -f target-ep-inv.yaml
    apictl apply -f target-ep-prod.yaml
    apictl apply -f target-ep-rev.yaml  

    Output:
    targetendpoint.wso2.com/inventory-ep created
    targetendpoint.wso2.com/products-ep created
    targetendpoint.wso2.com/review-ep created
    ```

 - Excute the following command to check if the backend services' pods are avaialable.
    ```
    >> apictl get pods

    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE
    inventory-ep-7f4fb4bb9b-hz6cq          1/1     Running   0          99s
    products-ep-fd6b4dbc8-wpf6n            1/1     Running   0          97s
    review-ep-7598f459ff-v5vpn             1/1     Running   0          73s
    ```
    
- Basic swagger definition belongs to the "OnlineStore" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
- In this swagger definition we have following resource paths and each of them use different backends.
    - products backend service
        - GET /products
        - GET /products/{productId}
    - inventory backend service
        - GET, POST /inventory/{productId}
    - reviews backend service
        - GET /review/{productId}
        
    Hence, in this swagger definition endpoint has been referred in resource level. 
    - The backend service of the "products" service has been mentioned as follows.
        ```
        /products:
            get:
              x-wso2-production-endpoints:
                urls:
                  - products-ep
              responses:
                "200":
                  description: ""
        
        /products/{productId}:
            get:
              parameters:
                - name: productId
                  in: path
                  required: true
                  schema:
                    type: string
              x-wso2-production-endpoints:
                urls:
                  - products-ep
              responses:
                "200":
                  description: ""
        ```
    - The backend service of the inventory service has been referred as follows.
    
        ```
        /inventory/{productId}:
            get:
              parameters:
                - name: productId
                  in: path
                  required: true
                  schema:
                    type: string
              x-wso2-production-endpoints:
                urls:
                  - inventory-ep
              responses:
                "200":
                  description: ""
            post:
              parameters:
                - name: productId
                  in: path
                  required: true
                  schema:
                    type: string
              x-wso2-production-endpoints:
                urls:
                  - inventory-ep
              requestBody:
                content:
                  application/json:
                    schema:
                      type: object
                      properties:
                        payload:
                          type: string
                description: Request Body
              responses:
                "200":
                  description: ""
        ```
    - The backend service of the review service has been referred as follows.
        
        ```
        /review/{productId}:
          get:
            parameters:
              - name: productId
                in: path
                required: true
                schema:
                  type: string
            x-wso2-production-endpoints:
              urls:
                - review-ep
            responses:
              "200":
                description: ""
        ```

- Deploy the API <br /> 
    ```
    >> apictl add api -n online-store-api --from-file=swagger.yaml --override
    
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/online-store-api-1-swagger created
    creating API definition
    api.wso2.com/online-store-api created
    ```
    
- Get available API <br /> 

    ```
    >> apictl get apis

    Output:    
    NAME               AGE
    online-store-api   17s
    ```

- Get service details to invoke the API<br />

    ```
    >> apictl get services

    Output:
    NAME               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    inventory-ep       ClusterIP      10.83.3.16     <none>           80/TCP                          57m
    online-store-api   LoadBalancer   10.83.1.223    35.232.188.134   9095:31850/TCP,9090:31981/TCP   13m
    products-ep        ClusterIP      10.83.6.91     <none>           80/TCP                          57m
    review-ep          ClusterIP      10.83.13.117   <none>           80/TCP                          56m
    ```
    - You can see all the backend(products, inventory & review) service and the managed API service(online-store-api) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
   - Invoking the products endpoint
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/storemep/v1.0.0/products" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (online-store-api), which then call its endpoint("products-ep" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
     {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
    - Invoking reviews endpoint
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/storemep/v1.0.0/review/151" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (online-store-api), which then call its endpoint("review-ep" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
    {"Review":{"productId":151, "reviewScore":8}}
    ```
    - Invoking inventory endpoint
  
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/storemep/v1.0.0/inventory/151" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (online-store-api), which then call its endpoint("inventory-ep" service) available in the cluster. If the request is success, you would be able to see the response as below.
    ```
    {"status":"SUCCESS","data":"false"}
    ```
- Delete the  API and Target Endpoint resources
    ```
    apictl delete api online-store-api
    apictl delete targetendpoints inventory-ep
    apictl delete targetendpoints products-ep
    apictl delete targetendpoints review-ep

    Output:
    api.wso2.com "online-store-api" deleted
    targetendpoint.wso2.com "inventory-ep" deleted
    targetendpoint.wso2.com "products-ep" deleted
    targetendpoint.wso2.com "review-ep" deleted
    ```