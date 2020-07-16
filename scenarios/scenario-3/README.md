## Scenario 3 - Deploy pet store service as a managed API secured with Basic Auth

- This scenario demonstrates how to deploy basic auth protected API in Kubernetes cluster.
- First we will create Kubernetes opaque secret with username and password of the user.
- Then we wil deploy "Security" custom resource(CR) for the basic auth by referring the secret created above in the kubernetes cluster.
- Created "Security" CR will be referred in the swagger definition.
- We deploy an API using that swagger file in the Kubernetes cluster.

***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics
 
 ##### Deploying the artifacts

- Navigate to the scenarios/scenario-3 directory and execute the following command

- Creating the opaque secret with basic credentials
    - In this scenario, we are using username "admin" and password "admin".
    ```$xslt
    >> apictl apply -f secret-basic.yaml
    
    Output:
    secret/secret-basic created
    ```
- Deploying "Security" custom resource in the k8s cluster.<br /> 
    ```$xslt
    >> apictl apply -f petstore-basic.yaml
    
    Output:
    security.wso2.com/petstorebasic created
    ```
    
- Refer the "Security" CR in the swagger file to indicate security of the API.
    ```$xslt
    security:
      - petstorebasic: []
    ```
- Deploy the API <br /> 
    - Following command with deploy the petstore service as a managed API in the k8s cluster.
    
    ```
    >> apictl add api -n petstore-basic --from-file=swagger.yaml --override
    
    Output:
    creating configmap with swagger definition
    configmap/petstore-basic-swagger created
    creating API definition
    api.wso2.com/petstore-basic created
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
- Check the API's service is deployed<br />
    ```
    >> apictl get services
    
    Output:
    NAME             TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)                         AGE
    petstore-basic   LoadBalancer   10.83.12.78   34.69.182.133   9095:30251/TCP,9090:30985/TCP   113s
    ```
    - You can see petstore-basic service is available for the use.
    - Get the external IP of the petstore-basic service
 
- Invoking the API
    - Prepare the basic credentials
        - Use base64 encode on "username:password" of the secret created at the beginning
            - Base64 encoded value of `admin:admin` -> `YWRtaW46YWRtaW4=`
                ```
                BASIC=YWRtaW46YWRtaW4=
                ```
   
    ```
    >> curl -X GET "https://<external IP of LB service>:9095/petstorebasic/v1/pet/55" -H "accept: application/xml" -H "Authorization:Basic $BASIC" -k
    ```    
    - Once you execute the above command, it will call to the managed API (petstore-basic), which then call its endpoint(https://petstore.swagger.io/v2). If the request is success, you would be able to see the response as below.
        ```
        <?xml version="1.0" encoding="UTF-8" standalone="yes"?><Pet><category><id>55</id><name>string</name></category><id>55</id><name>SRC_TIME_SIZE</name><photoUrls><photoUrl>string</photoUrl></photoUrls><status>available</status><tags><tag><id>55</id><name>string</name></tag></tags></Pet>
        ```
     
- Delete the  API <br /> 
    - Following command will delete all the artifacts created with this API including pods, deployment and services.

        ```
        >> apictl delete api petstore-basic
        
        Output:
        api.wso2.com "petstore-basic" deleted
        ```