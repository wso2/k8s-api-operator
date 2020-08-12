## Scenario 12 - Apply distributed rate-limiting to managed API in Kubernetes cluster
- This scenario describes how to apply distributed rate-limiting to a managed API in Kubernetes cluster
- Distributed rate limiting is required to handle throttling when there is more than 1 microgateway per API (i.e. when the number of replicas is greater than 1)
- First, we will deploy a rate-limiting custom resource which contains the policy ( 4 requests per minute)
- Then the created rate-limiting policy/CR will be referred in the swagger definition of the API
- Petstore service will be exposed as a managed API with ratelimiting in the Kubernetes cluster with 2 gateway replicas
- Finally, we will invoke the API continuously and observe the throttling behavior in 2 cases: 
    1. When distributed throttling is ***disabled***
    2. When distributed throttling is ***enabled***

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.


 ##### Deploying the artifacts

1. **When distributed throttling is disabled**

 - By default the distributed ratelimiting is disabled in controller-configs/controller_conf.yaml as below
    ```sh
    enabledGlobalTMEventPublishing: "false"
    ```
 
 - Navigate to `scenarios/scenario-12` directory.
  
 - Deploy the sample rate-limiting CR using the following command.
    ```sh
    >> apictl apply -f four-req-policy.yaml
    
    Output:
    ratelimiting.wso2.com/fourreqpolicy created
    ```

 - Prepared petstore basic swagger definition can be found within this directory.
 - Rate limiting policies to be applied on the API, should be mentioned in the swagger file with the "***x-wso2-throttling-tier***" extension.
In this swagger definition, the rate limiting policy has been mentioned as follows.
    ```sh
    x-wso2-throttling-tier: fourreqpolicy
    ```
 - Execute the following to expose pet-store as an API.

 - Create API with 2 replicas <br /> 
    ```sh
    >> apictl add api -n petstore-dist-rate --from-file=swagger.yaml --replicas=2
  
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/petstore-dist-rate-1-swagger created
    creating API definition
    api.wso2.com/petstore-dist-rate created
    ```
    
 - Get available API <br /> 
    ```sh
    >> apictl get apis
    
    Output:
    NAME                 AGE
    petstore-dist-rate   5m
    ```

 - Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```sh
    >> apictl get services
    
    Output:
    NAME                TYPE           CLUSTER-IP   EXTERNAL-IP       PORT(S)                         AGE
    petstore-dist-rate   LoadBalancer   10.83.4.44   104.197.114.248   9095:30680/TCP,9090:30540/TCP   8m20s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
 - Invoking the API <br />
    ```sh
    >> TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
    ```sh
    >> curl -X GET "https://<external IP of LB service>:9095/petstoredistrate/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
  
    - Once you execute the above command, it will call to the managed API (petstore-dist-rate), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```xml
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
    - From the ratelimiting policy we deployed earlier, we expect the API to allow only four requests per minute. 
    - But the microgateway uses an in-memory mechanism to handle throttling. Hence the rate limiting happens in the node-level, which results in multiplying the allowed limits when the microgateway is scaled.
    - Lets call the API with above curl for more than 8 times. After 8th time, you will get the following error message saying that the request has been throttled out.
        ```json
        {"fault":{"code":900802, "message":"Message throttled out", "description":"You have exceeded your quota"}}
        ```
    - Hence, in this scenario although we expect only 4 requests to be allowed, it actually allows 8 requests, since there are 2 gateways (2 replicas). Each gateway has allowed 4 requests.

2. **When distributed throttling is enabled**

    - To enable distributed rate limiting we need a central traffic management system. We will use the API Portal as the central traffic management system here.
        ***Important:***
        > Deploy the API Portal if you have not already deployed it. Refer the [Install the API portal and security token service](../../README.md#step-4-install-the-api-portal-and-security-token-service).

    - When distributed throttling is enabled, the API Microgateway upon recieving a request, checks against the local counter and if throttling limit  has not exceeded it publishes the events via a stream to a central traffic management solution. This is done over HTTP. The  central traffic management solution then  executes throttle policies against the events streams. When a particular request is throttled, the  central traffic management solution sends the details of the throttled out event to a JMS topic. Each API Microgateway node is subscribed to this JMS topic, and updates the local counter when the JMS topic is updated.  Hence the API Microgateway nodes gets notified of the throttle decisions through JMS messages.

    - Enable distributed rate limiting by modifying the apim-config in `controller-configs/controller_conf.yaml` as below. Set the "enabledGlobalTMEventPublishing" to "true". Default JMS port of the API Portal is used here.

        ```yaml
        #Enable distributed ratelimiting. Default value:false. If enabled please deploy API Portal
        enabledGlobalTMEventPublishing: "true"
        #The central traffic management solution URL (related to distributed ratelimiting)
        #Format: hostname_of_API_Portal:Default_port
        throttleEndpoint: "wso2apim.wso2:9443"
        #Message broker connection URL (related to distributed ratelimiting and token revocation)
        #Format: hostname_of_API_Portal:JMS_port
        jmsConnectionProvider: "wso2apim.wso2:5672"
    
        ```
   
    - Apply the changes
        ```sh
        >> apictl apply -f <k8s-api-operator-home>/api-operator/deploy/controller-configs/controller_conf.yaml
        ```
   
    - Delete the previous API if you had deployed it in earlier case
        ```sh
        >> apictl delete api petstore-dist-rate
        ```
   
    - Create API with 2 replicas with overriding previous docker image <br /> 
      **Note:** When you use the --override flag, it builds the docker image and pushes to the docker registry although it is available in the docker registry. If you are using AWS ECR as the registry type, delete the image of the API.
        ```sh
        >> apictl add api -n petstore-dist-rate --from-file=swagger.yaml --replicas=2 --override=true
        ``` 
   
    - Since the throttling is managed by the central traffic management system (i.e. API Portal), the same rate limiting policy should exist in the API Portal too.
    - Go to https://wso2apim:32001/admin Admin Portal and log in giving "admin" as username and password.
    - Create an advanced policy with the same name ("fourreqpolicy") and details as the policy you used earlier. Refer the below screenshot.

    ![Alt text](images/creating_ratelimit_policy.png?raw=true "Title")

    - Invoke the API as you did in the earlier case and observe that now only 4 requests are allowed per minute, and the 5th request is throttled out

- Delete the  API <br /> 
    Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```sh
    >> apictl delete api petstore-dist-rate
    >> apictl delete -f k8s-artifacts/api-portal/wso2-namespace.yaml
  
    Output:
    api.wso2.com "petstore-dist-rate" deleted
    namespace "wso2" deleted
    ```
    