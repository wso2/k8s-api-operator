## Scenario 19 - Deploy petstore service as a managed API secured with API Key Authentication
- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API secured with API key authentication. Hence the API invocation will be done with an API key only.  
- The WSO2 API Microgateway expects a self-contained JWT as an API Key.
- First we will deploy Security custom resource(CR) according to API key security related configurations. 
- Security CR created in the previous step should be referred in the Swagger definition using security extension.
- When API refers the Security CR in swagger definition under security keyword you need to make sure that the namespace of the Security CR is same as the namespace that the API belongs to.
- Then swagger definition will be deployed in the Kubernetes cluster as a managed API. 
 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 #### Secure APIs using API key Authentication
 ##### API Key Security Token Service Configuration

- API Key Security Token Service can be configured in the apim-config in `controller-configs/controller_conf.yaml`
    ```
    #APIKey issuer configurations
    #APIKey STS token configurations
    enabledAPIKeyIssuer: "true"
    apiKeyKeystorePath: "${mgw-runtime.home}/runtime/bre/security/ballerinaKeystore.p12"
    apiKeyKeystorePassword: "ballerina"
    apiKeyIssuerName: "https://localhost:9095/apikey"
    apiKeyIssuerCertificateAlias: "ballerina"
    validityTime: "-1"
    allowedAPIs: |
      - Petstore-Apikey: v1, v2, v3
    ```
- You can configure the validity you can configure a validity period for the API key token in validityTime. The default value is -1 which indicates unlimited time. 

- You can provide the list of allowed APIs by the generated API Key.
  ```
  allowedAPIs: |
    - API name given in the API Definition: Allowed versions of that API
  ```
- Apply the changes
   ```$xslt
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
    ```
 ##### Defining security schemes

- Security schemes can be defined on the swagger definition under securitySchemes. One or more API key security schemes can be used (as in logical OR) at the same time. A unique name for "name", query or header for "in"  and apiKey as "type" needs to be given for the defined API Key security scheme.

- If a security scheme is not defined it will use the default security scheme, "api_key" as the name of the header.

 ##### Deploying the artifacts
 
- Navigate to scenarios/scenario-19 directory.

- In the API key Security custom resource you can configure validateAllowedAPIs as true or false. If it is true when validating the API Key token, it only validates the allowed APIs of the API Key token which were provided in the controller_conf.yaml under allowedAPIs.

- Deploy the API key Security custom resource.
    ```$xslt
    >> apictl apply -f apiKey-security.yaml
    
    Output:
    security.wso2.com/petstoreapikey created
    ```

- Prepared petstore swagger definition can be found within this directory.

- Security schema of the API is referred in the swagger file with the "security" extension.
In this swagger definition, the security schema of the "petstore" service has been mentioned as follows.
    ```
    security:
      - petstoreapikey: []
    ```
    This can be referred either in the root level of the swagger(globally) or under resources such that the defined security will be reflected only for a specific resource.
- Execute the following to expose pet-store as a managed API.

- Deploy the  API <br /> 
    ```
    >> apictl add api -n petstore-apikey --from-file=swagger.yaml --override
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-apikey-swagger created
    creating API definition
    api.wso2.com/petstore-apikey created
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
    
- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME              AGE
    petstore-apikey   53m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME              TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
    petstore-apikey   LoadBalancer   10.8.15.95   104.198.161.233   9095:32367/TCP,9090:30323/TCP   53m

    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Obtain the API key token using the following command. <br />

    ```
    TOKEN=$(curl -X get "https://<EXTERNAL-IP>:9095/apikey" -H "Authorization:Basic YWRtaW46YWRtaW4=" -k)
    ```
- Invoking the API <br />
    ```
    >> curl -X GET "https://<EXTERNAL-IP of LB service>:9095/petstoreapikey/v1/pet/5" -H "accept: application/xml" -H "api_key:$TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API, which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
    - If the token is invalid or if the token is expired it will give an error response as below.
       ```
       {"fault":{"code":900901, "message":"Invalid Credentials", "description":"Invalid Credentials. Make sure you have given the correct access token"}}
       ```
- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
    ```
    >> apictl delete api petstore-apikey
    
    Output:
    api.wso2.com "petstore-apikey" deleted
    ```