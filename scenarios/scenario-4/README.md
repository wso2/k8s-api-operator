## Scenario 4 - Deploy petstore service as a managed API secured with JWT
- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API secured with JWT. Hence the API invocation will be done with a JWT token only.
- First, we will create a Kubernetes secret from the the public cert of the auth server(JWT token issuer).
- Then we will deploy Security custom resource(CR) according to JWT security related configurations including the above created secret. 
- Security CR created in the previous step should be referred in the Swagger definition's security extension as below.
- Created JWT Security will be referred in the swagger definition of the API.
- Final swagger definition will be deployed in the Kubernetes cluster as a managed API. 

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 ##### Deploying the artifacts
 
- Navigate to api-k8s-crds-1.1.0-alpha/scenarios/scenario-4 directory.

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
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/petstorejwt/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
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