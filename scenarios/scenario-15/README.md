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
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
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
  