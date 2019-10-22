## Scenario 4 - Deploy petstore service as a managed API secured with JWT
- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API secured with JWT. Hence the API invocation will be done with a JWT token only.
- First, we will create a Kubernetes secret from the the public cert of the auth server(JWT token issuer).
- Then we will deploy Security custom resource(CR) according to JWT security related configurations including the above created secret. 
- Security CR created in the previous step should be referred in the Swagger definition's security extension as below.
- Created JWT Security will be referred in the swagger definition of the API.
- Final swagger definition will be deployed in the Kubernetes cluster as a managed API. 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

- Navigate to wso2am-k8s-crds-1.0.0/scenarios/scenario-4 directory.

- Deploy Kubernetes secret of the public cert of the JWT token issues and JWT Security custom resource.
    ```$xslt
        apictl apply -f jwt-security.yaml
    ```
    - Output:
    ```$xslt
        security.wso2.com/petstorejwt created
        secret/wso2am300-secret created
    ```

- Prepared petstore swagger definition can be found within this directory.

- Security schema of the API is referred in the swagger file with the "security" extension.
In this swagger definition, the security schema of the "petstore" service has been mentioned as follows.
    ```
         security:
           - petstorejwt: []
    ```
    This can be referred either in the root level of the swagger(globally) or under resources such that the defined security will be reflected only for a specific resource.
- Execute the following to expose pet-store as a managed API.

- Deploy the  API <br /> 
    ```
        apictl add api -n petstore-jwt --from-file=swagger.yaml
    ``` 
    - Output:
    ```
        creating configmap with swagger definition
        configmap/petstore-jwt-swagger created
        api.wso2.com/petstore-jwt created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
        NAME          AGE
        petstore-jwt   3m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME           TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
        petstore-jwt   LoadBalancer   10.83.10.125   35.188.53.193   9095:32465/TCP,9090:30163/TCP   4m39s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjV9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06OTQ0M1wvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoiOFpWV1lQYkk2Rm1lY0ZoeXdVaDVVSXJaNEFvYSIsImV4cCI6MzcxODI5OTU1MiwiaWF0IjoxNTcwODE1OTA1LCJqdGkiOiJkMGI2NTgwNC05NDk3LTQ5ZjktOTcxNC01OTJmODFiNzJhYjMifQ==.HYCPxCbNcALcd0svu47EqFoxnnBAkVJSnCPnW6jJ1lZQTzSAiuiPcGzTnyP1JHodQknhYsSrvdZDIzWzU_mRH2i3-lMVdm0t43r-0Ti0EdBSX2756ilo266MVeWhxbz9p3hPm5ndDCoo_bfB4KbjigjmhXv_PJyUMuWtMo669sHQNs5FkiOT2X0gzFP1iJUFf-H9y762TEIYpylKedVDzQP8x4LCRZsO54e1iA-DZ5h5MKQhJsbKZZ_MMXGmtdo8refPyTCc7HIuevUXIWAaSNRFYj_HZTSRYhFEUtDWn_tJiySn2umRuP3XqxPmQal0SxD7JiV8DQxxyylsGw9k6g==
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
    ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
    ```
    

- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```
        apictl delete api petstore-jwt
    ```
    -  Output:
    ```
        api.wso2.com "petstore-jwt" deleted
    ```