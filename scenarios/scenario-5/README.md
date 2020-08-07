## Scenario 5 - Deploy petstore service as a managed API secured with OAuth2
- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API secured with OAuth2. Hence the API invocation will be done with a opaque OAuth2 token only.
- First, we will create Kubernetes opaque secret with username and password of the key manager server.
- Then, we will create another Kubernetes secret from the the public cert of the auth server(Key Manager).
- After that, we will deploy Security custom resource(CR) according to OAuth2 security related configurations including the above created secret. 
- Created OAuth2 Security CR will be referred in the swagger definition of the API.
- Then the petstore service will be deployed in the Kubernetes cluster as a managed API with OAuth2 secured.
- Then we will create the API and push it to API Manager deployment to obtain OAuth2 access token. **Hence we would need API Portal in the Kubernetes cluster to try out this scenario**. 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 ##### Deploying the artifacts

- Navigate to scenarios/scenario-5 directory.

- Deploy Kubernetes secret from the credentials of the key manager server.
    ```$xslt
    >> apictl apply -f credentials-secret.yaml

    Output:
    secret/oauth-credentials created
    ```

- Deploy Kubernetes secret of the public cert of the key manager server and OAuth2 Security custom resource.
    ```$xslt
    >> apictl apply -f oauth-security.yaml
  
    Output:
    security.wso2.com/petstoreoauth created
    secret/wso2am320-secret created
    ```

- Prepared petstore swagger definition can be found within this directory.

- Security schema of the API is referred in the swagger file with the "security" extension.
In this swagger definition, the security schema of the "petstore" service has been mentioned as follows.
    ```
    security:
      - petstoreoauth: []
    ```
    This can be referred either in the root level of the swagger(globally) or under resources such that the defined security will be reflected only for a specific resource.
- Execute the following to expose pet-store as a managed API.

- Deploy the  API <br /> 
    ```
    >> apictl add api -n petstore-oauth --from-file=swagger.yaml --override
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-oauth-swagger created
    api.wso2.com/petstore-oauth created
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.

- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME             AGE
    petstore-oauth   3m
    ```
   
- Obtain an access token to invoke the API
    - Publish the API in WSO2 API Manager deployment.
        - Add the APIM deployment as an environment to the apictl
            ```$xslt
            >> apictl add-env -e k8s --apim https://wso2apim:32001 --token https://wso2apim:32001/oauth2/token
            ```
        - Create the API project using swagger file with setting the initial state to `PUBLISHED`.
            ```$xslt
            >> apictl init petstore-oauth --oas=swagger.yaml --initial-state=PUBLISHED
            
            Output:
            Initializing a new WSO2 API Manager project in /home/wso2/k8s-api-operator/scenarios/scenario-5/petstore-oauth
            Project initialized
            Open README file to learn more
            ```
        - First line of the output shows the location of the API project.
        - Import the API to API Manager deployment
            ```$xslt
            >> apictl import-api -f petstore-oauth -e k8s -k
            
            Output:
            Login to k8s
            Username:admin
            Password:
            Logged into k8s environment
            WARNING: credentials are stored as a plain text in /Users/wso2/.wso2apictl/keys.json

            The specified API was not found.
            Creating: Petstore-Oauth v1
            Successfully imported API
            ```
        - Obtain OAuth2 access token
            ```$xslt
            >> apictl set --token-type oauth
          
            Output: 
            Token type set to:  oauth
            ```
            - Subscribe the API to to default application and get an access token using the following command.
                
                ```    
                >> apictl get-keys -n Petstore-Oauth -v v1 -r admin -k -e k8s
                
                Output: 
                API name:  Petstore-Oauth & version:  v1 exists
                API  Petstore-Oauth : v1 subscribed successfully.
                Access Token:  a68e6467-023e-3670-909c-11752449997e
                ```
- Invoking the API <br />

    - Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
        ```
        >> apictl get services
        
        Output:  
        NAME             TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
        petstore-oauth   LoadBalancer   10.83.10.125   35.188.53.193   9095:32465/TCP,9090:30163/TCP   4m39s
        ```
        - You can see petstore service has been exposed as a managed API.
        - Get the external IP of the managed API's service
         
            ```
            >> curl -X GET "https://<external IP of LB service>:9095/petstoreoauth/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer <Access-Token>" -k
            ```
    - Once you execute the above command, it will call to the managed API (petstore-oauth), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
    

- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
        ```
        >> apictl delete api petstore-oauth
        
        Output:
        api.wso2.com "petstore-oauth" deleted
        ```