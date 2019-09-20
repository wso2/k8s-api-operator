# k8s-apim-operator Scenarios

## Scenario 2

> ##### This scenario demonstrates Basic Authe protected API deployment and invocation

- Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
##### Navigate to the scenarios/scenario-2 directory and execute the following command

- Deploying the artifacts<br /> 
    - ***apimcli apply -f basic-security.yaml***
    - ***apimcli apply -f credential-secret.yaml***

- Create API <br /> 
    - ***apimcli add api -n petstore-basic --from-file=petstore_security_basic.yaml***

- Get service details to invoke the API<br />
    - ***apimcli get services***
    - Note: Get the external IP of the service
 
- Invoking the API <br />
    - ***BASIC=YWRtaW46YWRtaW4=***
   
    - ***curl -X GET "https://\<external IP of LB service>:9095/petstore/v1/pet/55" -H "accept: application/xml" -H "Authorization:Basic $BASIC" -k***

- Delete API <br /> 
    - ***apimcli delete api petstore-basic***
