# k8s-apim-operator Scenarios

## Scenario 4

> ##### This scenario demonstrates API deployment and invocation with OAuth security

- Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
##### Navigate to the scenarios/scenario-4 directory and execute the following command

- Deploying secrets and security kind for oauth (Todo: include apim endpoint and cert to security)

- Create API <br /> 
    - ***apimcli add api -n petstore-oauth --from-file=petstore_security_oauth.yaml***

- Update API <br /> 
    - ***apimcli update api -n petstore-oauth --from-file=petstore_security_oauth.yaml***
    
- Get available API <br /> 
    - ***apimcli get apis***

- Publish the API in the API Manager deployment<br> subcribe to an application and get an opaque oauth token.
    - Using the APIM CLI command, adding the environment to the CLI configs/
        ```
        apimcli add-env -e k8s --registration https://wso2apim:9443/client-registration/v0.15/register --apim https://wso2apim:9443 --token https://wso2apim:8243/token --admin https://wso2apim:9443/api/am/admin/v0.15 --api_list https://wso2apim:9443/api/am/publisher/v0.15/apis --app_list https://wso2apim:9443/api/am/store/v0.15/applications
        
        ```
    - Init the API project using CLI command

        ```
        apimcli init petstore --oas=./deploy/scenarios/scenario-1/petstore_basic.yaml
        ```

    - Import the API to the k8s environment.
(You need to change the API life cycle status before importing, to published in the api.yaml file to publish the API)
        ```
        ./apimcli import-api -f petstore/ -e k8s -k 
        ```

- Get service details to invoke the API<br />
    - ***apimcli get services***
    - Note: Get the external IP of the service
 
- Invoking the API <br />
    - Get an OAuth2 Opaque access token to invoke the API (make sure to include APIM endpoint and certificate of KM server in the security respectively)
   
    - ***curl -X GET "https://\<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $opaqueToken" -k***

- Delete API <br /> 
    - ***apimcli delete api petstore-oauth***
