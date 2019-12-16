## Scenario 12 - Apply distributed rate-limiting to managed API in Kubernetes cluster
- This scenario describes how to apply distributed rate-limiting to a managed API in Kubernetes cluster
- Distributed rate limiting is required to handle throttling when there is more than 1 microgateway per API (i.e. when the number of replicas is greater than 1)
- First, we will deploy a rate-limiting custom resource which contains the policy ( 4 requests per minute)
- Then the created rate-limiting policy/CR will be referred in the swagger definition of the API
- Petstore service will be exposed as a managed API with ratelimiting in the Kubernetes cluster with 2 gateway replicas
- Finally, we will invoke the API continuously and observe the throttling behavior in 2 cases: 
1. When distributed throttling is disabled
2. When distributed throttling is enabled

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts

1. **When distributed throttling is disabled**

 - By default the distributed ratelimiting is disabled in controller-configs/controller_conf.yaml as below

    ```
        enabledGlobalTMEventPublishing: "false"
    ```
 
 - Navigate to api-k8s-crds-1.0.1/scenarios/scenario-12 directory.
  
 - Deploy the sample rate-limiting CR using the following command.
    ```
        apictl apply -f four-req-policy.yaml
    ```
    - Output:
    ```
        ratelimiting.wso2.com/fourreqpolicy created
    ```

- Prepared petstore basic swagger definition can be found within this directory.
- Rate limiting policies to be applied on the API, should be mentioned in the swagger file with the "x-wso2-throttling-tier" extension.
In this swagger definition, the rate limiting policy has been mentioned as follows.
    ```
        x-wso2-throttling-tier: fourreqpolicy
    ```
- Execute the following to expose pet-store as an API.

- Create API with 2 replicas <br /> 
    ```
        apictl add api -n petstore-dist-rate --from-file=swagger.yaml --replicas=2
    ``` 
    - Output:
    ```
        creating configmap with swagger definition
        configmap/petstore-dist-rate-swagger created
        api.wso2.com/petstore-dist-rate created
    ```
    
- Get available API <br /> 
    ```
        apictl get apis
    ```
    - Output:
    ```    
        NAME                 AGE
        petstore-dist-rate   5m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
        apictl get services
    ```
    - Output:
    
    ```
        NAME                TYPE           CLUSTER-IP   EXTERNAL-IP       PORT(S)                         AGE
        petstore-dist-rate   LoadBalancer   10.83.4.44   104.197.114.248   9095:30680/TCP,9090:30540/TCP   8m20s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
        TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjMsInV1aWQiOm51bGx9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06MzIwMDFcL29hdXRoMlwvdG9rZW4iLCJ0aWVySW5mbyI6e30sImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6IjNGSWlUM1R3MWZvTGFqUTVsZjVVdHVTTWpsUWEiLCJleHAiOjM3MTk3Mzk4MjYsImlhdCI6MTU3MjI1NjE3OSwianRpIjoiZDI3N2VhZmUtNTZlOS00MTU2LTk3NzUtNDQwNzA3YzFlZWFhIn0.W0N9wmCuW3dxz5nTHAhKQ-CyjysR-fZSEvoS26N9XQ9IOIlacB4R5x9NgXNLLE-EjzR5Si8ou83mbt0NuTwoOdOQVkGqrkdenO11qscpBGCZ-Br4Gnawsn3Yw4a7FHNrfzYnS7BZ_zWHPCLO_JqPNRizkWGIkCxvAg8foP7L1T4AGQofGLodBMtA9-ckuRHjx3T_sFOVGAHXcMVwpdqS_90DeAoT4jLQ3darDqSoE773mAyDIRz6CAvNzzsWQug-i5lH5xVty2kmZKPobSIziAYes-LPuR-sp61EIjwiKxnUlSsxtDCttKYHGZcvKF12y7VF4AqlTYmtwYSGLkXXXw
    ```
   
    ```
        curl -X GET "https://<external IP of LB service>:9095/petstoredistrate/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-dist-rate), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
    ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
    ```
    - From the ratelimiting policy we deployed earlier, we expect the API to allow only four requests per minute. 
    - But the microgateway uses an in-memory mechanism to handle throttling. Hence the rate limiting happens in the node-level, which results in multiplying the allowed limits when the microgateway is scaled.
    - Lets call the API with above curl for more than 8 times. After 8th time, you will get the following error message saying that the request has been throttled out.
    ```$xslt
        {"fault":{"code":900802, "message":"Message throttled out", "description":"You have exceeded your quota"}}
    ```
    - Hence, in this scenario although we expect only 4 requests to be allowed, it actually allows 8 requests, since there are 2 gateways (2 replicas). Each gateway has allowed 4 requests.

2. **When distributed throttling is enabled**

    - To enable distributed rate limiting we need a central traffic management system. We will use the API Portal as the central traffic management system here.

     ***Important:***
    > Deploy the API Portal if you have not already deployed it, using \<api-k8s-crds-home>/k8s-artifacts

    - When distributed throttling is enabled, the API Microgateway upon recieving a request, checks against the local counter and if throttling limit  has not exceeded it publishes the events via a stream to a central traffic management solution. This is done over HTTP. The  central traffic management solution then  executes throttle policies against the events streams. When a particular request is throttled, the  central traffic management solution sends the details of the throttled out event to a JMS topic. Each API Microgateway node is subscribed to this JMS topic, and updates the local counter when the JMS topic is updated.  Hence the API Microgateway nodes gets notified of the throttle decisions through JMS messages.

    - Enable distributed rate limiting by modifying the apim-config in controller-configs/controller_conf.yaml as below. Set the "enabledGlobalTMEventPublishing" to "true". Default JMS port of the API Portal is used here.

    ```
    #Enable distributed ratelimiting. Default value:false. If enabled please deploy API Portal
    enabledGlobalTMEventPublishing: "true"
    #The central traffic management solution URL (related to distributed ratelimiting)
    #Format: hostname_of_API_Portal:Default_port
    throttleEndpoint: "wso2apim.wso2:32001"
    #Message broker connection URL (related to distributed ratelimiting and token revocation)
    #Format: hostname_of_API_Portal:JMS_port
    jmsConnectionProvider: "wso2apim.wso2:28230"

    ```
    - Apply the changes
    ```
        kubectl apply -f <api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml
    ```
    - Delete the previous API if you had deployed it in earlier case
    ```
        apictl delete api petstore-dist-rate
    ```
    - Create API with 2 replicas <br /> 
    ```
        apictl add api -n petstore-dist-rate --from-file=swagger.yaml --replicas=2 --override=true
    ``` 
    - Since the throttling is managed by the central traffic management system (i.e. API Portal), the same rate limiting policy should exist in the API Portal too.
    - Go to https://wso2apim/admin Admin Portal and log in giving "admin" as username and password.
    - Create an advanced policy with the same name and details as the policy you used earlier. Refer the below screenshot.

    ![Alt text](images/creating_policy.png?raw=true "Title")

    - Invoke the API as you did in the earlier case and observe that now only 4 requests are allowed per minute, and the 5th request is throttled out

- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```
        apictl delete api petstore-dist-rate
    ```
    -  Output:
    ```
        api.wso2.com "petstore-dist-rate" deleted
    ```
    