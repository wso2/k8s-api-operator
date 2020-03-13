## Scenario 6 - Apply rate-limiting to managed API in Kubernetes cluster
- This scenario describes how to apply rate-limiting to a managed API in Kubernetes cluster
- First, we will deploy a rate-limiting custom resource which contains the policy ( 5 requests per minute)
- Then the created rate-limiting policy/CR will be referred in the swagger definition of the API
- Petstore service will be exposed as a managed API with ratelimiting in the Kubernetes cluster 
- Finally, we will invoke the API continuously and observe the throttling behavior.

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts
 
 - Navigate to api-k8s-crds-1.1.0-alpha/scenarios/scenario-6 directory.
  
 - Deploy the sample rate-limiting CR using the following command.
    ```
        apictl apply -f five-req-policy.yaml
    ```
    - Output:
    ```
        ratelimiting.wso2.com/fivereqpolicy created
    ```

- Prepared petstore basic swagger definition can be found within this directory.
- Rate limiting policies to be applied on the API, should be mentioned in the swagger file with the "x-wso2-throttling-tier" extension.
In this swagger definition, the rate limiting policy has been mentioned as follows.
    ```
        x-wso2-throttling-tier: fivereqpolicy
    ```
- Execute the following to expose pet-store as an API.

- Create API <br /> 
    ```
        apictl add api -n petstore-rate --from-file=swagger.yaml
    ``` 
    - Output:
    ```
        creating configmap with swagger definition
        configmap/petstore-rate-swagger created
        api.wso2.com/petstore-rate created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
        NAME          AGE
        petstore-rate   5m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME            TYPE           CLUSTER-IP   EXTERNAL-IP       PORT(S)                         AGE
        petstore-rate   LoadBalancer   10.83.4.44   104.197.114.248   9095:30680/TCP,9090:30540/TCP   8m20s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/petstorerate/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-rate), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
    ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
    ```
    - Continue to call the API with above curl for more than 5 times. After 5th time, you will get the following error message saying that the request has been throttled out.
    ```$xslt
        {"fault":{"code":900802, "message":"Message throttled out", "description":"You have exceeded your quota"}}
    ```

- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```
        apictl delete api petstore-rate
    ```
    -  Output:
    ```
        api.wso2.com "petstore-rate" deleted
    ```
    