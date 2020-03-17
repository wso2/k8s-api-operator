## Scenario 15 - Apply java interceptors to an API

- This scenario describes how to apply interceptors written in java as .jar files to carry out transformations and mediations on the requests and responses.
- First, we need to implement custom request interceptors and response interceptors. We have provided sample .jar file in scenario-15. If you want to learn more about implementing custom java interceptors you can refer the document [adding interceptors.](https://docs.wso2.com/display/MG310/Message+Transformation)
- Then we need to Initialize a new API project and add the .jar files in libs folder.
- We need to refer the interceptors in swagger definition in order to apply them on the requests and responses.
- Finally, we will invoke the API and observe how the added interceptors act on requests and responses.

***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Init the API project using CLI. This will Initialize a new API project in same directory.
 
     ```
     apictl init petstore-int --oas=interceptor_swagger.yaml
    
    Output:    
    Initializing a new WSO2 API Manager project in ./product-apim-tooling/import-export-cli/build/target/apimcli/petstore-int
    Project initialized
    Open README file to learn more
     ```
 - Copy the _mgw-interceptor.jar_ file in scenario-15 into the libs folder in petstore-int/libs path.
  
    ***Note:***  
    > In the above interceptor we have defined a function _interceptRequest_, which validates whether the request has the header "X-API-KEY" and a function _interceptResponse_ send a custom json message if the response contains the key "error". You can find more information [here.](https://docs.wso2.com/display/MG310/Message+Transformation#0057f1e771984fca9b6964fe0e1e1937)

- Java Interceptors can be added to a particular resource or to the whole API as needed. We use OpenAPI extensions to refer interceptors in swagger definition.
     - java interceptor consists with class  _org.wso2.micro.gateway.interceptor.SampleInterceptor_ which intercept the request and response flows. This will refer in the swagger definition as follow.
     
        ```
        x-wso2-request-interceptor: java:org.wso2.micro.gateway.interceptor.SampleInterceptor
        x-wso2-response-interceptor: java:org.wso2.micro.gateway.interceptor.SampleInterceptor
        ```
- Execute the following to expose pet-store as an API.

- Create the API

    ```
    apictl add api -n petstore-int --from-file=petstore-int
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-int-swagger created
    creating configmap with interceptors
    configmap/petstore-int-interceptors created
    api.wso2.com/petstore-int created
    ```
- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)

    ```
    apictl get services
    
    Output:
    NAME            TYPE           CLUSTER-IP   EXTERNAL-IP       PORT(S)                         AGE
    petstore-int   LoadBalancer   10.83.4.44   104.197.114.248   9095:30680/TCP,9090:30540/TCP   8m20s
    ```
    
    - You can see petstore-int service has been exposed as a managed API.
    - Get the external IP of the managed API's service

- Invoking the API

    ```
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
   
    ```
    curl -X GET "https://<External_IP>:9095/petstoreint/v1/pet/55"  -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    
    - Once you execute the above command, it will call to the managed API (petstore-int), which then call its endpoint(https://petstore.swagger.io/v2). Since the request header did not contain "X-API-KEY", you would be able to see the error response as below.
    
    ```
    {"error":"Missing required header"}
    ```

    - Then invoke the API with an "X-API-KEY" header as follows.
    
     ```
     curl -X GET "https://<External_IP>:9095/petstore/v1/pet/55" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -H "X-API-KEY: 6fa741de1bdd1d91830ba" -k -v
     ```
            
    - Once you execute the above command, you will get the success response message as follows.
                
    ```json
    {"id":55, "category":{"id":55, "name":"string"}, "name":"SRC_TIME_SIZE", "photoUrls":["string"], "tags":[{"id":55, "name":"string"}], "status":"available"}
    ```

- Delete the API

    ```
    apictl delete api petstore-int
    ``` 
  
  - Output
   ```
    api.wso2.com "petstore-int" deleted
    ``` 
  