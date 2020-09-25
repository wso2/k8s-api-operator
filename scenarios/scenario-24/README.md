## Scenario 24 - Using event hub in WSO2 API Manager
- This scenario describes how to connect to WSO2 API Manager as event hub.
- When you enable connecting to WSO2 API Manager as event hub, micro gateway will connect to API Manager and retrieve the
API and subscriptions related data from API Manager.
- By connecting to WSO2 API Manager as event hub you can validate JWT subscriptions.
- You need to run WSO2 API manager where there is access from Micro gateway. In this scenario we are using 
Kubernetes cluster to deploy WSO2 API Manager. You can follow the steps mentioned in [here](https://github.com/wso2/K8s-api-operator#step-4-install-the-api-portal-and-security-token-service) to deploy
WSO2 API Manager on Kubernetes.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.
  
 ##### Deploying the artifacts
 
- Enabling for connecting to WSO2 API Manager as event hub for retrieving API and subscription data from API Manager
can be configured in the apim-config in `controller-configs/controller_conf.yaml`.
- You can set `enabledEventhub` value to `true`. 
- You can configure API Manager URL by providing the `throttleEndpoint` and the message broker connection URL by providing
`jmsConnectionProvider` values.
     ```
     # Enable configurations for retrieving API and subscription data from API Manager.
     enabledEventhub: "true"
     #Format: hostname_of_API_Portal:Default_port
     throttleEndpoint: "wso2apim.wso2:9443"
     #Format: hostname_of_API_Portal:JMS_port
     jmsConnectionProvider: "wso2apim.wso2:5672"
     ```
- You can configure User name and password as base64 encoded for APIM (The internal data API) in the apim-secret in `controller-configs/controller_conf.yaml`.
     ```
     apiVersion: v1
     kind: Secret
     metadata:
        name: apim-secret
        namespace: wso2-system
     type: Opaque
     data:
        #Base64 encoded username and password for APIM
        username: YWRtaW4=
        password: YWRtaW4=
     ```
  
- Apply the changes
    ```$xslt
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
    ```
  
- Then we will create the API and push it to API Manager deployment to obtain JWT access token. Hence we would need API Portal in the Kubernetes cluster to try out this scenario.

- Navigate to `scenarios/scenario-24` directory.

- Deploy Kubernetes secret of the public cert of the JWT token issues and JWT Security custom resource.
- In this JWT Security custom resource the `validateSubscription` field is set to true to validate JWT subscriptions.
    ```$xslt
    >> apictl apply -f jwt-security-sub.yaml
    
    Output:
    security.wso2.com/petstorejwt created
    secret/wso2am320-secret created
    ```
Note: ***audience*** field can be provided in Security custom resource(CR) under securityConfig field, if you have any.

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
    >> apictl add api -n petstore-jwt --from-file=swagger.yaml --override
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-jwt-swagger created
    creating API definition
    api.wso2.com/petstore-jwt created
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
    
- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME           INITIAL-REPLICAS   MODE         ENDPOINT        AGE
    petstore-jwt   1                  privateJet   10.106.28.187   20m

    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME           TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
    petstore-jwt   LoadBalancer   10.83.10.125   10.106.28.187   9095:32465/TCP,9090:30163/TCP   4m39s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service

- Add the API portal as an environment to the API controller using the following command.
     ```sh
      >> apictl add-env -e k8s \
                  --apim https://wso2apim:32001 \
                  --token https://wso2apim:32001/oauth2/token
      
      Output:
      Successfully added environment 'k8s'
     ```
  
- Initialize the API project using API Controller
     ```sh
      >> apictl init petstore-api \
                  --oas=swagger.yaml \
                  --initial-state=PUBLISHED
      
      Output:
      Initializing a new WSO2 API Manager project in scenarios/scenario-24/petstore-api
      Project initialized
      Open README file to learn more
     ```
  
- Import the API to the API portal. **[IMPORTANT]**
      For testing purpose use ***admin*** as username and password when prompted.
      </br>
     ```sh
      >> apictl login k8s -k
      >> apictl import-api -f petstore-api/ -e k8s -k
      
      Output:
      Successfully imported API
     ```
  
- By default the API is secured with JWT. Hence a valid JWT token is needed to invoke the API.
  You can obtain a JWT token using the API Controller command as below.
     ```sh
      >> apictl get-keys -n Petstore-Jwt -v v1 -e k8s --provider admin -k
  
**Note:** You also have the option to generate a token by logging into the devportal,
creating an application, subscribing to an API and generating JWT token. 

 
- Invoking the API <br />
    ```
    TOKEN= <Access Token Value>
    ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstorejwt/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
   

- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```
    >> apictl delete api petstore-jwt
    
    Output:
    api.wso2.com "petstore-jwt" deleted
    ```
