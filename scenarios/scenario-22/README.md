## Scenario 22 - Generating Backend JWT

- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on 
a kubernetes cluster as a managed API while using JWT Generation in API Micro-gateway to send 
a customized JWT to the backend with user preferred claims.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

#### Setting the header of the backend request

- When you are passing a token to the backend or when you are generating a JWT, you can configure the header in which the JWT token will be passed to the backend. 
You can configure the name of the header in the apim-config in `controller-configs/controller_conf.yaml`. By default, the header is X-JWT-Assertion.

    ```
    #JWT header when forwarding the request to the backend
    jwtHeader: "X-JWT-Assertion"
    ```
#### Generating a JWT using default configurations

- You can customize the token generation from the configurations included in the apim-config in `controller-configs/controller_conf.yaml` without using 
a custom JWT generator, since a default JWT generator implementation is included.

    ```
      #JWT Generator configurations
      #Enable jwt generator
      enabledJwtGenerator: "true"
      #Dialect prefix that can be added to the claims
      jwtClaimDialect: "http://wso2.org/claims"
      #Signature algorithm used to sign the JWT token (only SHA256withRSA and NONE is supported)
      jwtSigningAlgorithm: "SHA256withRSA"
      #Certificate alias from the keystore
      jwtCertificateAlias: "ballerina"
      #Private key alias from the keystore
      jwtPrivateKeyAlias: "ballerina"
      #JWT token expiry time - ms (valid only if the jwt generator caching mechanism is disabled)
      jwtTokenExpiry: "900000"
      #Restricted claims as a list that should not be included in the backend JWT token
      jwtRestrictedClaims: |
      # claim1  (This is an example)
    
      #Token issuer standard claim
      jwtIssuer: "wso2.org/products/am"
      #Token audience standard claim as a list
      jwtAudience: |
      # http://org.wso2.apimgt/gateway (This is an example)
    
      #JWT token generator implementation
      jwtGeneratorImpl: "org.wso2.micro.gateway.jwt.generator.MGWJWTGeneratorImpl"
      #JWT Generator cache configurations
      #Enable jwt generator token caching
      jwtTokenCacheEnable: "true"
      #Token cache expiry time (ms)
      jwtTokenCacheExpiryTime: "900000"
      #Token cache capacity
      jwtTokenCacheCapacity: "10000"
      #Token cache eviction factor
      jwtTokenCacheEvictionFactor: "0.25"
    ```
- The properties and the customizations of the generated JWT token can be configured and also the caching aspects of the generated JWT tokens can be 
configured.

- Apply the changes
   ```$xslt
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
    ```
 
- Create API <br /> 
    ```
    >> apictl add api -n petstore-api --from-file=swagger.yaml --override
        
    Output:
    creating configmap with swagger definition
    configmap/petstore-api-swagger created
    creating API definition
    api.wso2.com/petstore-api created
    ``` 
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
    
- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME           INITIAL-REPLICAS   MODE         ENDPOINT        AGE
    petstore-api   1                  privateJet   34.67.188.203   2m33s
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME           TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)                         AGE
    petstore-api   LoadBalancer   10.8.9.30    34.67.188.203   9095:32202/TCP,9090:31588/TCP   36s

    ```
  
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
- When you are invoking a resource with a JWT token, you will be able to obtain the generated backend JWT.
You can obtain the generated backend JWT token by enabling the TRACE log level.

- Delete the  API
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
        ```
        >> apictl delete api petstore-api
        
        Output:
        api.wso2.com "petstore-api" deleted
        ```
      
#### Generating a JWT using a custom JWT Generator

- Micro-gateway defines an AbstractMGWJWTGenerator abstract class to write a custom JWT generator class. 
The developers can extend this abstract class to implement their logic to generate the JWT. 
There are two abstract methods named populateStandardClaims and populateCustomClaims where you can write your logic to populate claims in the JWT. 
Furthermore, you have the capability to write your own logic overriding any of the current methods in the abstract class.

##### Adding a custom JWT generator to the project

- After the JWT generator is written, that JWT generator project should be built and you can obtain a jar.
- In this scenario we have provided a .jar file generated like that.
- Then we need to Initialize a new API project and add the .jar files in libs folder.

##### Deploying the artifacts

- You can provide the classpath of the custom JWT generator in the generatorImpl configuration in the apim-config in `controller-configs/controller_conf.yaml`.

    ```
    #JWT token generator implementation
    jwtGeneratorImpl: "sample.jwt.generator.SampleJWTGenerator"
    ```
- Apply the changes
   ```$xslt
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
    ```

- Init the API project using CLI. This will Initialize a new API project in same directory.
    ```
    >> apictl init petstore-int --oas=swagger.yaml
    
    Output:    
    Initializing a new WSO2 API Manager project in k8s-api-operator-1.2.2/scenarios/scenario-22/petstore-int
    Project initialized
    ```
- Copy the _sample-jwt-generator-1.0-SNAPSHOT.jar_ file in scenario-15 into the libs folder in petstore-int/libs path.
     ```
      >> cp sample-jwt-generator-1.0-SNAPSHOT.jar petstore-int/libs/
     ```
- Create API <br /> 
    ```
      >> apictl add api -n petstore-api --from-file=petstore-int --override
          
      Output:
      creating configmap with swagger definition
      configmap/petstore-api-1-swagger created
      creating configmap with java interceptor petstore-api-1-8081-jar-intcpt
      configmap/petstore-api-1-8081-jar-intcpt created
      creating API definition
      api.wso2.com/petstore-api created
    ``` 
      Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
      
- Get available API <br /> 
    ```
      >> apictl get apis
      
      Output:
      NAME           INITIAL-REPLICAS   MODE         ENDPOINT        AGE
      petstore-api   1                  privateJet   35.226.223.89   2m30s
    ```
- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
      >> apictl get services
      
      Output:
      NAME           TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)                         AGE
      petstore-api   LoadBalancer   10.8.8.30    35.226.223.89   9095:32342/TCP,9090:32676/TCP   48s

    ```
  
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
   
- Invoking the API <br />
     ```
      TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
     
    ```
      >> curl -X GET "https://<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
    ```
  
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
         ```
          <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
         ```

  - When you are invoking a resource with a JWT token, you will be able to obtain the generated backend JWT.
  You can obtain the generated backend JWT token by enabling the TRACE log level.
  
- Delete the  API
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
          
      ```
          >> apictl delete api petstore-api
          
          Output:
          api.wso2.com "petstore-api" deleted
      ```
        