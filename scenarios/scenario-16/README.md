## Scenario 16 - Deploy multiple swagger-projects as one API (Shared Mode)

- This scenario describes how to expose multiple APIs as single API gateway on a kubernetes cluster as a managed API.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

#### Deploying the artifacts

- Navigate to scenarios/scenario-16 directory.
- Prepared petstore swagger definitions `pets_swagger.yaml` and `stores_swagger.yaml` can be found within this directory.
- Base paths of the APIs are mentioned in the swagger files with the "x-wso2-basetpath" vendor extension.
    
    Base path in `pets_swagger.yaml` file
    ```
    x-wso2-basePath: /pet/{version}
    ```
  
    Base path in `stores_swagger.yaml` file
    ```
    x-wso2-basePath: /store/{version}
    ```

- Init API projects using CLI. This will Initialize a new API project in same directory.

    ```sh
    >> apictl init pets-int --oas=pets_swagger.yaml
    
    Output:    
    Initializing a new WSO2 API Manager project in .../scenarios/scenario-16/pets-int
    Project initialized
    Open README file to learn more
    ```
    ```sh
    >> apictl init stores-int --oas=stores_swagger.yaml
    
    Output:    
    Initializing a new WSO2 API Manager project in .../scenarios/scenario-16/stores-int
    Project initialized
    Open README file to learn more
    ```

- Create API

    ```sh
    >> apictl add api -n petstore-multiple-api --from-file=pets-int --from-file=stores-int

    Output:
    Processing swagger 1: pets-int
    creating configmap with swagger definition
    configmap/petstore-multiple-api-1-swagger created
    Processing swagger 2: stores-int
    creating configmap with swagger definition
    configmap/petstore-multiple-api-2-swagger created
    creating API definition
    api.wso2.com/petstore-multiple-api created
    ```
  
    **Optional Parameters**
    
    ```
    --mode=privatejet   Overrides the deploying mode. Available modes: privateJet, sidecar
    --version=2.0.0     Used for docker image versioning. Default value is v1.0.0

    >> apictl add api -n petstore-multiple-api --from-file=pets-int --from-file=stores-int --mode=privatejet --version=2.0.0
    ```
    
- Get available APIs

    ```
    >> apictl get apis

    Output:   
    NAME                    AGE
    petstore-multiple-api   57s
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    
    ```
    >> apictl get services

    Output:
    NAME                    TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
    petstore-multiple-api   LoadBalancer   10.106.24.198   localhost     9095:30029/TCP,9090:32027/TCP   2m14s  
    ```
    - You can see petstore-multiple-api service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API

    ```sh
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
    **Invoke Pets API**
    
    ```sh
    >> curl -X GET "https://<external IP of LB service>:9095/pet/v1/pet/1" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
  
    If the output message is "Pet not found" try with different pet id.
  
    Output:
    
    ```json
    {"id":10,"category":{"id":10,"name":"dolor"},"name":"eiusmod","photoUrls":["${photoUrls}","${photoUrls}"],"tags":[{"id":10000,"name":"Lorem"},{"id":10000,"name":"Lorem"}],"status":"consectetur"}
    ```
  
    **Invoke Stores API**
    
    ```sh
    >> curl -X GET "https://<external IP of LB service>:9095/store/v1/store/inventory" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
  
    Output:
    
    ```json
    {"incididunt":4,"tempor":6,"string":626,"pending":19,"adipiscing":6,"available":109,"do":6,"dolor":4,"freaky":2,"sed":4,"scary":1,"sit":10,"ut":8,"sold":8,"labore":2,"eiusmod":10,"magna":6,"et":4,"0":5,"dolore":6,"for sale":1,"Lorem":4,"amet":8,"ipsum":10,"elit":2,"consectetur":8}
    ```
    
- Delete the API

    Following command will delete all the artifacts created with this API including pods, deployment and services.
    
    ```sh
    >> apictl delete api petstore-multiple-api

    Output:
    api.wso2.com "petstore-multiple-api" deleted
    ```