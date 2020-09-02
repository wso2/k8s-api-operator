## Scenario 4 - Deploy petstore service as a managed API secured with JWT
- This scenario describes how to deploy the petstore service(https://petstore.swagger.io/v2) on a kubernetes cluster as a managed API secured with JWT. Hence the API invocation will be done with a JWT token only. 
- There can be use cases, where multiple jwt issuers or key managers are used. In that case operator can be configured to work with JWTs issued by all of them.
- This scenario describes configuring Multiple JWT issuers as well.  
- First, we will create a Kubernetes secret from the the public certificate(s) of the oauth server(s)(JWT token issuer(s)).
- Then we will deploy Security custom resource(CR) according to JWT security related configurations including the above created secret. 
- Security CR created in the previous step should be referred in the Swagger definition using security extension.
- Then swagger definition will be deployed in the Kubernetes cluster as a managed API. 
 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 #### Secure APIs using JWT (Self Contained JWT)
  
 ##### Deploying the artifacts
 
- Navigate to scenarios/scenario-4 directory.

- Deploy Kubernetes secret of the public cert of the JWT token issues and JWT Security custom resource.
  - Create the **certificate secret** with base64 encoded **PEM** format of public certificate the key **server.pem**.
    Or the **certificate secret** can be created with the following command.
    ```shell script
    >> apictl create secret generic <CERT_SECRET_NAME> --from-file=server.pem=<PUBLIC_CERT_PEM_FORMAT>
    ```
    
    We have created a **certificate secret** and a Security CRD for you which can be found in the file
    [jwt-security.yaml](jwt-security.yaml) that we are going to apply.
    
    ```$xslt
    >> apictl apply -f jwt-security.yaml
    
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
    NAME          AGE
    petstore-jwt   3m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME           TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
    petstore-jwt   LoadBalancer   10.83.10.125   35.188.53.193   9095:32465/TCP,9090:30163/TCP   4m39s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br />
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
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
  
 #### Configure Multiple JWT issuers
 
 ##### Deploying the artifacts
  
 - Navigate to scenarios/scenario-4 directory.
 
 - Deploy Kubernetes secrets of the public certificates of the JWT token issuers and JWT Security custom resource.
     ```$xslt
     >> apictl apply -f multiple-jwt-security.yaml
     
     Output:
     security.wso2.com/petstorejwt created
     secret/wso2am320-secret created
     secret/jwt-secret created
     ```
- Created security CR name should be refer in the swagger definition under security as follows.
    ```
    security:
      - petstorejwt: []
    ```
    This can be referred either in the root level of the swagger(globally) or under resources such that the defined security will be reflected only for a specific resource.
    
- Prepared petstore swagger definition can be found within this directory.

- Execute the following to expose petstore as a managed API.

- Deploy the  API <br /> 
    ```
    >> apictl add api -n petstore-multiple-jwt --from-file=swagger.yaml --override
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-multiple-jwt-swagger created
    creating API definition
    api.wso2.com/petstore-multiple-jwt created
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
- Get available API <br /> 
    ```
    >> apictl get apis
    
    Output:
    NAME          AGE
    petstore-multiple-jwt   3m
    ```

- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME           TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                         AGE
    petstore-multiple-jwt   LoadBalancer   10.83.10.125   35.188.53.193   9095:32465/TCP,9090:30163/TCP   4m39s
    ```
    - You can see petstore service has been exposed as a managed API.
    - Get the external IP of the managed API's service
 
- Invoking the API <br /> The API can be invoke with any of the JWT tokens taken by configured JWT issuers.
    
    - Invoking the API with token obtained from configured trusted JWT issuer at the micro gateway.
    ```
    TOKEN1=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
    ```
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstorejwt/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN1" -k
    ```
    - Other JWT issuer used in this example is https://dev-5uc4e1dm.au.auth0.com/. the access token can be obtained with following curl command and can execute the API as follows.
    ```
    >> curl --request POST \
          --url https://dev-5uc4e1dm.au.auth0.com/oauth/token \
          --header 'content-type: application/json' \
          --data '{"client_id":"sBK5eEaN5PjuBTaMz3SHgUGpV405ALrX","client_secret":"esg9eMQ7TFuxC8dAB3LqmGcpuKOUi604UgmxkLOAolSzyKfpsFJh24Y-Wpn-rfW7","audience":"https://dev-5uc4e1dm.au.auth0.com/api/v2/","grant_type":"client_credentials"}'
    
    Ouput:
    {
       "access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlEwVXhRekl6UlRsQ05qRXhRa05FTWtZeE1EWTNSVVJFUmtKQ00wWkVPVFpFT1VNM01UZzRRdyJ9.eyJpc3MiOiJodHRwczovL2Rldi01dWM0ZTFkbS5hdS5hdXRoMC5jb20vIiwic3ViIjoic0JLNWVFYU41UGp1QlRhTXozU0hnVUdwVjQwNUFMclhAY2xpZW50cyIsImF1ZCI6Imh0dHBzOi8vZGV2LTV1YzRlMWRtLmF1LmF1dGgwLmNvbS9hcGkvdjIvIiwiaWF0IjoxNTg1OTgzMTU4LCJleHAiOjE1ODYwNjk1NTgsImF6cCI6InNCSzVlRWFONVBqdUJUYU16M1NIZ1VHcFY0MDVBTHJYIiwic2NvcGUiOiJyZWFkOmNsaWVudF9ncmFudHMgY3JlYXRlOmNsaWVudF9ncmFudHMgZGVsZXRlOmNsaWVudF9ncmFudHMgdXBkYXRlOmNsaWVudF9ncmFudHMgcmVhZDp1c2VycyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnVzZXJzX2FwcF9tZXRhZGF0YSB1cGRhdGU6dXNlcnNfYXBwX21ldGFkYXRhIGRlbGV0ZTp1c2Vyc19hcHBfbWV0YWRhdGEgY3JlYXRlOnVzZXJzX2FwcF9tZXRhZGF0YSByZWFkOnVzZXJfY3VzdG9tX2Jsb2NrcyBjcmVhdGU6dXNlcl9jdXN0b21fYmxvY2tzIGRlbGV0ZTp1c2VyX2N1c3RvbV9ibG9ja3MgY3JlYXRlOnVzZXJfdGlja2V0cyByZWFkOmNsaWVudHMgdXBkYXRlOmNsaWVudHMgZGVsZXRlOmNsaWVudHMgY3JlYXRlOmNsaWVudHMgcmVhZDpjbGllbnRfa2V5cyB1cGRhdGU6Y2xpZW50X2tleXMgZGVsZXRlOmNsaWVudF9rZXlzIGNyZWF0ZTpjbGllbnRfa2V5cyByZWFkOmNvbm5lY3Rpb25zIHVwZGF0ZTpjb25uZWN0aW9ucyBkZWxldGU6Y29ubmVjdGlvbnMgY3JlYXRlOmNvbm5lY3Rpb25zIHJlYWQ6cmVzb3VyY2Vfc2VydmVycyB1cGRhdGU6cmVzb3VyY2Vfc2VydmVycyBkZWxldGU6cmVzb3VyY2Vfc2VydmVycyBjcmVhdGU6cmVzb3VyY2Vfc2VydmVycyByZWFkOmRldmljZV9jcmVkZW50aWFscyB1cGRhdGU6ZGV2aWNlX2NyZWRlbnRpYWxzIGRlbGV0ZTpkZXZpY2VfY3JlZGVudGlhbHMgY3JlYXRlOmRldmljZV9jcmVkZW50aWFscyByZWFkOnJ1bGVzIHVwZGF0ZTpydWxlcyBkZWxldGU6cnVsZXMgY3JlYXRlOnJ1bGVzIHJlYWQ6cnVsZXNfY29uZmlncyB1cGRhdGU6cnVsZXNfY29uZmlncyBkZWxldGU6cnVsZXNfY29uZmlncyByZWFkOmhvb2tzIHVwZGF0ZTpob29rcyBkZWxldGU6aG9va3MgY3JlYXRlOmhvb2tzIHJlYWQ6ZW1haWxfcHJvdmlkZXIgdXBkYXRlOmVtYWlsX3Byb3ZpZGVyIGRlbGV0ZTplbWFpbF9wcm92aWRlciBjcmVhdGU6ZW1haWxfcHJvdmlkZXIgYmxhY2tsaXN0OnRva2VucyByZWFkOnN0YXRzIHJlYWQ6dGVuYW50X3NldHRpbmdzIHVwZGF0ZTp0ZW5hbnRfc2V0dGluZ3MgcmVhZDpsb2dzIHJlYWQ6c2hpZWxkcyBjcmVhdGU6c2hpZWxkcyBkZWxldGU6c2hpZWxkcyByZWFkOmFub21hbHlfYmxvY2tzIGRlbGV0ZTphbm9tYWx5X2Jsb2NrcyB1cGRhdGU6dHJpZ2dlcnMgcmVhZDp0cmlnZ2VycyByZWFkOmdyYW50cyBkZWxldGU6Z3JhbnRzIHJlYWQ6Z3VhcmRpYW5fZmFjdG9ycyB1cGRhdGU6Z3VhcmRpYW5fZmFjdG9ycyByZWFkOmd1YXJkaWFuX2Vucm9sbG1lbnRzIGRlbGV0ZTpndWFyZGlhbl9lbnJvbGxtZW50cyBjcmVhdGU6Z3VhcmRpYW5fZW5yb2xsbWVudF90aWNrZXRzIHJlYWQ6dXNlcl9pZHBfdG9rZW5zIGNyZWF0ZTpwYXNzd29yZHNfY2hlY2tpbmdfam9iIGRlbGV0ZTpwYXNzd29yZHNfY2hlY2tpbmdfam9iIHJlYWQ6Y3VzdG9tX2RvbWFpbnMgZGVsZXRlOmN1c3RvbV9kb21haW5zIGNyZWF0ZTpjdXN0b21fZG9tYWlucyByZWFkOmVtYWlsX3RlbXBsYXRlcyBjcmVhdGU6ZW1haWxfdGVtcGxhdGVzIHVwZGF0ZTplbWFpbF90ZW1wbGF0ZXMgcmVhZDptZmFfcG9saWNpZXMgdXBkYXRlOm1mYV9wb2xpY2llcyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlcyBkZWxldGU6cm9sZXMgdXBkYXRlOnJvbGVzIHJlYWQ6cHJvbXB0cyB1cGRhdGU6cHJvbXB0cyByZWFkOmJyYW5kaW5nIHVwZGF0ZTpicmFuZGluZyByZWFkOmxvZ19zdHJlYW1zIGNyZWF0ZTpsb2dfc3RyZWFtcyBkZWxldGU6bG9nX3N0cmVhbXMgdXBkYXRlOmxvZ19zdHJlYW1zIGNyZWF0ZTpzaWduaW5nX2tleXMgcmVhZDpzaWduaW5nX2tleXMgdXBkYXRlOnNpZ25pbmdfa2V5cyIsImd0eSI6ImNsaWVudC1jcmVkZW50aWFscyJ9.kXX_MfZ8wPY5s9B94xdudJivcHNYS2SY5YI63jmnqAQpru7uWkeYgNNW8Tkq2BO1o8c-oNjWbNksm2iSioU1qBJWDhRT81i-9qn1y9xGMWBKk7Wg0WNvyQni9iRHGyc3tdiYJb1YeXjTxBe2gbpLLNVUAZKOWYhsUh0CKd3LRO1DrJJK_DkpO3D7oOxdLpPI3g_xEQWZhRX3VZ3FKudqKE-rgClL4m62wJGy70PNnyhpn6TppjrmzpSQX0xOuW3qVhoj9Y0wNOtVo5lXJdCuGKozbfcM7JTFBMIPJ_a7uWT7Cc8Uio7LfCYqVAMfevGDVUWXxhPpR90bjC5aDxaL2A",
       "scope":"read:client_grants create:client_grants delete:client_grants update:client_grants read:users update:users delete:users create:users read:users_app_metadata update:users_app_metadata delete:users_app_metadata create:users_app_metadata read:user_custom_blocks create:user_custom_blocks delete:user_custom_blocks create:user_tickets read:clients update:clients delete:clients create:clients read:client_keys update:client_keys delete:client_keys create:client_keys read:connections update:connections delete:connections create:connections read:resource_servers update:resource_servers delete:resource_servers create:resource_servers read:device_credentials update:device_credentials delete:device_credentials create:device_credentials read:rules update:rules delete:rules create:rules read:rules_configs update:rules_configs delete:rules_configs read:hooks update:hooks delete:hooks create:hooks read:email_provider update:email_provider delete:email_provider create:email_provider blacklist:tokens read:stats read:tenant_settings update:tenant_settings read:logs read:shields create:shields delete:shields read:anomaly_blocks delete:anomaly_blocks update:triggers read:triggers read:grants delete:grants read:guardian_factors update:guardian_factors read:guardian_enrollments delete:guardian_enrollments create:guardian_enrollment_tickets read:user_idp_tokens create:passwords_checking_job delete:passwords_checking_job read:custom_domains delete:custom_domains create:custom_domains read:email_templates create:email_templates update:email_templates read:mfa_policies update:mfa_policies read:roles create:roles delete:roles update:roles read:prompts update:prompts read:branding update:branding read:log_streams create:log_streams delete:log_streams update:log_streams create:signing_keys read:signing_keys update:signing_keys",
       "expires_in":86400,
       "token_type":"Bearer"
    }
        
    ```
    Use the ***access_token*** value in the above response you get as the value of the TOKEN2 shown below.
    ```
    TOKEN2=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlEwVXhRekl6UlRsQ05qRXhRa05FTWtZeE1EWTNSVVJFUmtKQ00wWkVPVFpFT1VNM01UZzRRdyJ9.eyJpc3MiOiJodHRwczovL2Rldi01dWM0ZTFkbS5hdS5hdXRoMC5jb20vIiwic3ViIjoic0JLNWVFYU41UGp1QlRhTXozU0hnVUdwVjQwNUFMclhAY2xpZW50cyIsImF1ZCI6Imh0dHBzOi8vZGV2LTV1YzRlMWRtLmF1LmF1dGgwLmNvbS9hcGkvdjIvIiwiaWF0IjoxNTg1MjE5MDI0LCJleHAiOjE1ODUzMDU0MjQsImF6cCI6InNCSzVlRWFONVBqdUJUYU16M1NIZ1VHcFY0MDVBTHJYIiwic2NvcGUiOiJyZWFkOmNsaWVudF9ncmFudHMgY3JlYXRlOmNsaWVudF9ncmFudHMgZGVsZXRlOmNsaWVudF9ncmFudHMgdXBkYXRlOmNsaWVudF9ncmFudHMgcmVhZDp1c2VycyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnVzZXJzX2FwcF9tZXRhZGF0YSB1cGRhdGU6dXNlcnNfYXBwX21ldGFkYXRhIGRlbGV0ZTp1c2Vyc19hcHBfbWV0YWRhdGEgY3JlYXRlOnVzZXJzX2FwcF9tZXRhZGF0YSByZWFkOnVzZXJfY3VzdG9tX2Jsb2NrcyBjcmVhdGU6dXNlcl9jdXN0b21fYmxvY2tzIGRlbGV0ZTp1c2VyX2N1c3RvbV9ibG9ja3MgY3JlYXRlOnVzZXJfdGlja2V0cyByZWFkOmNsaWVudHMgdXBkYXRlOmNsaWVudHMgZGVsZXRlOmNsaWVudHMgY3JlYXRlOmNsaWVudHMgcmVhZDpjbGllbnRfa2V5cyB1cGRhdGU6Y2xpZW50X2tleXMgZGVsZXRlOmNsaWVudF9rZXlzIGNyZWF0ZTpjbGllbnRfa2V5cyByZWFkOmNvbm5lY3Rpb25zIHVwZGF0ZTpjb25uZWN0aW9ucyBkZWxldGU6Y29ubmVjdGlvbnMgY3JlYXRlOmNvbm5lY3Rpb25zIHJlYWQ6cmVzb3VyY2Vfc2VydmVycyB1cGRhdGU6cmVzb3VyY2Vfc2VydmVycyBkZWxldGU6cmVzb3VyY2Vfc2VydmVycyBjcmVhdGU6cmVzb3VyY2Vfc2VydmVycyByZWFkOmRldmljZV9jcmVkZW50aWFscyB1cGRhdGU6ZGV2aWNlX2NyZWRlbnRpYWxzIGRlbGV0ZTpkZXZpY2VfY3JlZGVudGlhbHMgY3JlYXRlOmRldmljZV9jcmVkZW50aWFscyByZWFkOnJ1bGVzIHVwZGF0ZTpydWxlcyBkZWxldGU6cnVsZXMgY3JlYXRlOnJ1bGVzIHJlYWQ6cnVsZXNfY29uZmlncyB1cGRhdGU6cnVsZXNfY29uZmlncyBkZWxldGU6cnVsZXNfY29uZmlncyByZWFkOmhvb2tzIHVwZGF0ZTpob29rcyBkZWxldGU6aG9va3MgY3JlYXRlOmhvb2tzIHJlYWQ6ZW1haWxfcHJvdmlkZXIgdXBkYXRlOmVtYWlsX3Byb3ZpZGVyIGRlbGV0ZTplbWFpbF9wcm92aWRlciBjcmVhdGU6ZW1haWxfcHJvdmlkZXIgYmxhY2tsaXN0OnRva2VucyByZWFkOnN0YXRzIHJlYWQ6dGVuYW50X3NldHRpbmdzIHVwZGF0ZTp0ZW5hbnRfc2V0dGluZ3MgcmVhZDpsb2dzIHJlYWQ6c2hpZWxkcyBjcmVhdGU6c2hpZWxkcyBkZWxldGU6c2hpZWxkcyByZWFkOmFub21hbHlfYmxvY2tzIGRlbGV0ZTphbm9tYWx5X2Jsb2NrcyB1cGRhdGU6dHJpZ2dlcnMgcmVhZDp0cmlnZ2VycyByZWFkOmdyYW50cyBkZWxldGU6Z3JhbnRzIHJlYWQ6Z3VhcmRpYW5fZmFjdG9ycyB1cGRhdGU6Z3VhcmRpYW5fZmFjdG9ycyByZWFkOmd1YXJkaWFuX2Vucm9sbG1lbnRzIGRlbGV0ZTpndWFyZGlhbl9lbnJvbGxtZW50cyBjcmVhdGU6Z3VhcmRpYW5fZW5yb2xsbWVudF90aWNrZXRzIHJlYWQ6dXNlcl9pZHBfdG9rZW5zIGNyZWF0ZTpwYXNzd29yZHNfY2hlY2tpbmdfam9iIGRlbGV0ZTpwYXNzd29yZHNfY2hlY2tpbmdfam9iIHJlYWQ6Y3VzdG9tX2RvbWFpbnMgZGVsZXRlOmN1c3RvbV9kb21haW5zIGNyZWF0ZTpjdXN0b21fZG9tYWlucyByZWFkOmVtYWlsX3RlbXBsYXRlcyBjcmVhdGU6ZW1haWxfdGVtcGxhdGVzIHVwZGF0ZTplbWFpbF90ZW1wbGF0ZXMgcmVhZDptZmFfcG9saWNpZXMgdXBkYXRlOm1mYV9wb2xpY2llcyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlcyBkZWxldGU6cm9sZXMgdXBkYXRlOnJvbGVzIHJlYWQ6cHJvbXB0cyB1cGRhdGU6cHJvbXB0cyByZWFkOmJyYW5kaW5nIHVwZGF0ZTpicmFuZGluZyByZWFkOmxvZ19zdHJlYW1zIGNyZWF0ZTpsb2dfc3RyZWFtcyBkZWxldGU6bG9nX3N0cmVhbXMgdXBkYXRlOmxvZ19zdHJlYW1zIGNyZWF0ZTpzaWduaW5nX2tleXMgcmVhZDpzaWduaW5nX2tleXMgdXBkYXRlOnNpZ25pbmdfa2V5cyIsImd0eSI6ImNsaWVudC1jcmVkZW50aWFscyJ9.LfbVoGNZqAnAXFCbqjZ0SiBh_M_XbZigWry5wNxy8CPTdxeVtsh5eC1v1f3bNGNKkjMmlPkXBCBfkWNfqBrisdAKEvzGnyehQ-SvaiZcySLxT6cmiqvubAzDkHCsDyfkdEEnhc0lhGTJlAOi57Npxqi1snUcQGVR7GVNY7zu1gCFKgEMtNa4k8B7nX7-_fDGX4iCynNE49dYm7eEQyAJk7IeiHbCPEoqub_CGT8F6lOWq-Q75lkkwVxnwFfGZgD60gJP07l4SGquadQrbyWtSgKBZKLUsZVz8ibX2cP19JQKIn_wB2LOhJne-qmTPN2RPhX8yiPK-ZtTOnoQviFm5A
    ```
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstorejwt/v1/pet/55" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN2" -k
    ```
    - Once you execute the above command, it will call to the managed API (petstore-api), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
 - Delete the  API <br /> 
     - Following command will delete all the artifacts created with this API including pods, deployment and services.
         ```
         >> apictl delete api petstore-multiple-jwt
         
         Output:
         api.wso2.com "petstore-multiple-jwt" deleted
         ```