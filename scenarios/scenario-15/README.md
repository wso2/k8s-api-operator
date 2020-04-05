## Scenario 15 - Apply Java interceptors to an API

- This scenario describes how to apply interceptors written in Java as .jar files to carry out transformations and mediations on the requests and responses.
- First, we need to implement custom request interceptors and response interceptors. We have provided sample .jar file in scenario-15. If you want to learn more about implementing custom java interceptors you can refer the document [adding interceptors.](https://docs.wso2.com/display/MG310/Message+Transformation)
- Then we need to Initialize a new API project and add the .jar files in libs folder.
- We need to refer the interceptors in swagger definition in order to apply them on the requests and responses.
- Finally, we will invoke the API and observe how the added interceptors act on requests and responses.

***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


#### Deploying the artifacts
 
- Init the API project using CLI. This will Initialize a new API project in same directory.

    ```
    >> apictl init petstore-int --oas=swagger.yaml
    
    Output:    
    Initializing a new WSO2 API Manager project in ./product-apim-tooling/import-export-cli/build/target/apimcli/petstore-int
    Project initialized
    Open README file to learn more
    ```
- Copy the _mgw-interceptor.jar_ file in scenario-15 into the libs folder in petstore-int/libs path.

    ```
    >> cp mgw-interceptor.jar petstore-int/libs/
    ```

    ***Note:***  
    > In the above interceptor we have defined a function _interceptRequest_, which validates whether the request has the header "X-API-KEY" and a function _interceptResponse_ send a custom json message if the response contains the key "error". You can find more information [here.](https://docs.wso2.com/display/MG310/Message+Transformation#0057f1e771984fca9b6964fe0e1e1937)

- Java Interceptors can be added to a particular resource or to the whole API as needed. We use OpenAPI extensions to refer interceptors in swagger definition.
- Java interceptor consists with class  _org.wso2.micro.gateway.interceptor.SampleInterceptor_ which intercept the request and response flows. This will refer in the swagger definition as follow.
 
    ```
    x-wso2-request-interceptor: java:org.wso2.micro.gateway.interceptor.SampleInterceptor
    x-wso2-response-interceptor: java:org.wso2.micro.gateway.interceptor.SampleInterceptor
    ```
 
- Create the API

    ```
    >> apictl add api -n petstore-java-int --from-file=petstore-int
    
    Output:
    Processing swagger 1: petstore-int
    creating configmap with swagger definition
    configmap/petstore-java-int-1-swagger created
    creating configmap with java interceptor petstore-java-int-1-mgw-interceptor.jar
    configmap/petstore-java-int-1-mgw-interceptor.jar created
    creating API definition
    api.wso2.com/petstore-java-int created
    ```
- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)

    ```
    >> apictl get services
    
    Output:
    NAME            TYPE           CLUSTER-IP   EXTERNAL-IP       PORT(S)                         AGE
    petstore-int   LoadBalancer   10.83.4.44   104.197.114.248   9095:30680/TCP,9090:30540/TCP   8m20s
    ```
    
    - You can see petstore-int service has been exposed as a managed API.
    - Get the external IP of the managed API's service

- Invoking the API

    ```
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    ```
   
    ```
    >> curl -X GET "https://<External_IP>:9095/petstore/v1/pet/55"  -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -k
    ```
    
    - Once you execute the above command, it will call to the managed API (petstore-int), which then call its endpoint(https://petstore.swagger.io/v2). Since the request header did not contain "X-API-KEY", you would be able to see the error response as below.
    
    ```
    {"error":"Missing required header"}
    ```

    - Then invoke the API with an "X-API-KEY" header as follows.
    
     ```
     >> curl -X GET "https://<External_IP>:9095/petstore/v1/pet/55" -H "accept: application/json" -H "Authorization:Bearer $TOKEN" -H "X-API-KEY: 6fa741de1bdd1d91830ba" -k
     ```
            
    - Once you execute the above command, you will get the success response message as follows.
                
    ```json
    {"id":55, "category":{"id":55, "name":"string"}, "name":"SRC_TIME_SIZE", "photoUrls":["string"], "tags":[{"id":55, "name":"string"}], "status":"available"}
    ```
    **Note:** If the response message is "Pet not found" try with different pet id.

- Delete the API

    ```
    >> apictl delete api petstore-java-int
    
    Output:
    api.wso2.com "petstore-java-int" deleted
    ```
  