## Scenario 2 - Deploy pet store service as a managed API in k8s cluster

- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

##### Deploying the artifacts

- Navigate to scenarios/scenario-2 directory.
- Prepared petstore basic swagger definition can be found within this directory.
- Base path of the API should be mentioned in the swagger file with the  "x-wso2-basePath".
    ```
    x-wso2-basePath: /petstore/v1
    ```

- In this swagger definition, the backend service of the "petstore" service has been mentioned as follows. It can be either globally or resource level.
In the scenarios, we have defined it in resource level.

    ```
    x-wso2-production-endpoints:
      urls:
        - https://petstore.swagger.io/v2
    ```
- Execute the following to expose pet-store as an API.

- Create API <br /> 
    ```
    >> apictl add api -n petstore-api --from-file=swagger.yaml --override
        
    Output:
    creating configmap with swagger definition
    configmap/petstore-api-swagger created
    creating API definition
    api.wso2.com/petstore-api created
    ``` 
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
    
- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME          AGE
    petstore-api   5m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME                 TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)                         AGE
    petstore-api         LoadBalancer   10.83.14.186   35.188.53.193     9095:32264/TCP,9090:32540/TCP   2m42s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
    

- Delete the  API
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
        ```
        >> apictl delete api petstore-api
        
        Output:
        api.wso2.com "petstore-api" deleted
        ```