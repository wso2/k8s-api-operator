## Scenario 7 - Deploy APIs in k8s in private jet mode
- This scenario describes how to expose a service as a managed API in private jet mode.
- In private jet mode, backend and the managed API will be deployed in two different pods.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the `private jet` mode in the swagger file.
- Later we will deploy the API using the swagger definition 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to scenarios/scenario-7 directory and deploy the sample backend service using the following command.
    ```
    >> apictl apply -f products-privatejet.yaml

    Output:
    targetendpoint.wso2.com/products-privatejet created
    ```
- Basic swagger definition belongs to the "products" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
The mode of managed API (private jet or sidecar) also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "products" service and the managed API mode have been mentioned as follows.
    ```
    x-wso2-production-endpoints:
      urls:
        - products-privatejet
    x-wso2-mode: privateJet
    ```

- Create API <br /> 
    ```
    >> apictl add api -n products-pj --from-file=swagger.yaml --override

    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/products-pj-swagger created
    api.wso2.com/products-pj created
    ```
  
  **Note:** When you use the --override flag, it builds the docker image and pushes to the docker registry although it is available in the docker registry. If you are using AWS ECR as the registry type, delete the image of the API.
    
- Get available API <br /> 

    ```
    >> apictl get apis

    Output:   
    NAME          AGE
    products-pj   3m
    ```

- Get service details to invoke the API<br />
    ```
    >> apictl get services

    Output:
    NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
    products-pj           LoadBalancer   10.99.134.132   <pending>     9095:32290/TCP,9090:30057/TCP   1s
    products-privatejet   ClusterIP      10.101.34.213   <none>        80/TCP                          45m
    ```
    - You can see both the backend(products-privatejet) service and the managed API service(product-pj) is available.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/storepj/v1/products" -H "Authorization:Bearer $TOKEN" -k

    Output:
    {"products":[{"name":"Apples", "id":101, "price":"$1.49 / lb"}, {"name":"Macaroni & Cheese", "id":151, "price":"$7.69"}, {"name":"ABC Smart TV", "id":301, "price":"$399.99"}, {"name":"Motor Oil", "id":401, "price":"$22.88"}, {"name":"Floral Sleeveless Blouse", "id":501, "price":"$21.50"}]}
    ```
    
- List the pods and check how the backend services and the managed API have been deployed

    ```
    >> apictl get pods        

    Output:
    products-pj-699d65df7f-qt2vv           1/1     Running   0          5m12s
    products-privatejet-6777d6f5bc-gqfg4   1/1     Running   0          25m
    products-privatejet-6777d6f5bc-k88sl   1/1     Running   0          25m
    ```
- Delete the API and the sample backend service (Target Endpoint resource)
    ```
    >> apictl delete api products-pj
    >> apictl delete targetendpoints products-privatejet

    Output:
    api.wso2.com "products-pj" deleted
    targetendpoint.wso2.com "products-privatejet" deleted
    ```
