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
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
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