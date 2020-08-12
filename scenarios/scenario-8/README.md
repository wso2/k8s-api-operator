## Scenario 8 - Deploy APIs in k8s in sidecar mode
- This scenario describes how to expose a service as a managed API in sidecar mode.
- In sidecar mode, backend and the managed API will be deployed in the same pod.
- First we will deploy a target endpoint resource containing the information of the backend service
- Then we would refer the backend in the swagger file and set the `sidecar` mode in the swagger file.
- Later we will deploy the API using the swagger definition 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to scenarios/scenario-8 directory and deploy the sample backend service using the following command.
    ```
    >> apictl apply -f inventory-sidecar.yaml
   
    Output:
    targetendpoint.wso2.com/inventory-sidecar created
    ```
- Basic swagger definition belongs to the "inventory" service is available in swagger.yaml.
Backend endpoint of the API should be mentioned in the swagger file with the "x-wso2-production-endpoints" extension.
The mode of managed API (private jet or sidecar) also has to be mentioned in the swagger with the "x-wso2-mode" extension.
In this swagger definition, the backend service of the "products" service and the managed API mode have been mentioned as follows.
    ```
    x-wso2-production-endpoints:
      urls:
        - inventory-sidecar
    x-wso2-mode: sidecar
    ```

- Create API <br /> 
    ```
    >> apictl add api -n inventory-sc --from-file=swagger.yaml --override

    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/inventory-sc-1-swagger created
    creating API definition
    api.wso2.com/inventory-sc created
    ```
    **Note:** When you use the --override flag, it builds the docker image and pushes to the docker registry although it is available in the docker registry. If you are using AWS ECR as the registry type, delete the image of the API.
        
- Get available API <br /> 
    ```
    >> apictl get apis
  
    Output:
    NAME           AGE
    inventory-sc   3m
    ```

- Get service details to invoke the API<br />
    ```
    >> apictl get services

    Output:
    NAME                TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)                         AGE
    inventory-sc        LoadBalancer   10.109.107.234   <pending>     9095:31565/TCP,9090:31420/TCP   87s
    ```
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/storesc/v1/inventory/301" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (inventory-sc), which then call its endpoint ("inventory-sc-service" service) available in the same pod. If the request is success, you would be able to see the response as below.
    ```
    {"status":"SUCCESS","data":"true"}
    ```
    
- List the pods and check how the backend services and the managed API have been deployed

    ```
    >> apictl get pods

    Output:
    NAME                                    READY   STATUS    RESTARTS   AGE
    inventory-sc-664db4fd5f-xf84l           2/2     Running   0          7m14s
    ```
    - To list the containers running inside the pod
    ```
    >> apictl get pods POD_NAME_HERE -o jsonpath='{.spec.containers[*].name}'
    
    Replace the POD_NAME_HERE with the corresponding pod name

    Output:
    inventory-sc-service mgwinventory-sc
    ```
    - Here 
        - inventory-sc-service is the backend service container
        - mgwinventory-sc is the managed inventory API container
- Delete the  API and the Target Endpoint resource
    ```
    >> apictl delete api inventory-sc
    >> apictl delete targetendpoints inventory-sidecar

    Output:
    api.wso2.com "inventory-sc" deleted
    targetendpoint.wso2.com "inventory-sidecar" deleted
    ```
