# k8s-apim-operator

##### Navigate to the k8s-apim-operator/apim-operators/ directory and execute the following command

##### Deploy k8s client extensions
- Give executable permission to the extension file <br /> 
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-add.sh***
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-update.sh***
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-show.sh***
- Copy the extensions to ***/usr/local/bin/***
    - ___cp ./deploy/kubectl-extension/kubectl-add.sh /usr/local/bin___
    - ___cp ./deploy/kubectl-extension/kubectl-update.sh /usr/local/bin___
    - ___cp ./deploy/kubectl-extension/kubectl-show.sh /usr/local/bin___

###### Note:
- User may have to remove '.sh' from the extension depending on the kubernetes environment. Use the following additional commands for that.
    - mv /usr/local/bin/kubectl-add.sh /usr/local/bin/kubectl-add
    - mv /usr/local/bin/kubectl-update.sh /usr/local/bin/kubectl-update
    - mv /usr/local/bin/kubectl-show.sh /usr/local/bin/kubectl-show
##### Deploy k8s CRD artifacts

> ##### Before deploying the role you have to make yourself as a cluster admin. (Replace "email-address" with the proper value)
- *kubectl create clusterrolebinding email-address --clusterrole=cluster-admin --user=email-address*

> ##### Deploying namespace, roles/role binding and service account associated with operator
- _for i in ./deploy/controller-artifacts/*yaml; do kubectl apply -f $i; done_

> ##### Deploying CRD for API, Target endpoint, Security, Ratelimiting
- _for i in ./deploy/crd/*yaml; do kubectl apply -f $i; done_


> ##### Deploying controller level configuration
>> "controller-configs" contains the configuration user would have to change. Modify the controller_conf.yaml with the needed values and enter the ***user's docker registry***. Enter the base 64 encoded username and password of the user's docker registry into the docker_secret_template.yaml.
- _for i in ./deploy/controller-configs/*yaml; do kubectl apply -f $i; done_

> ##### Deploy sample custom resources on the kubernetes cluster (Optional)
- _for i in ./deploy/sample-crs/*yaml; do kubectl apply -f $i; done_

> ##### Deploying an API in K8s cluter

- Download sample API definition (swagger files) from [product micro-gateway sample](https://github.com/wso2/product-microgateway/tree/master/samples) github location.
- Execute the following command to deploy the API
    - *kubectl add api "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" -n="desired namespace"*
    - ex: ***kubectl add api petstorebasic --from-file=petstore_basic.yaml --replicas=2 -n=wso2***

- Execute the following command to update the API
    - *kubectl update api "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" -n="desired namespace"*
    - ex: ***kubectl update api petstorebasic --from-file=petstore_basic.yaml --replicas=1 -n=wso2***

- Execute the following command to view the details of the API
    - *kubectl show api "api_name" -n="desired namespace"*
    - ex: ***kubectl show api petstorebasic -n=wso2***

- Execute the following command to remove the API
    - *kubectl delete api "api_name" -n="desired namespace"*
    - ex: ***kubectl delete api petstorebasic -n=wso2***

- Execute the following command to remove the apim operator
    - ***kubectl delete deployment apim-operator -n=wso2-system***
    

- Note:
> Namespace and replicas are optional parameters. If they are not provided default namespace will be used and 1 replica will be created. However, the namespace used in all the commands related to particular API name must match.


##### Incorporating analytics to the k8s operator

- To enable analytics, modify the analytics-config configmap given in the ./deploy/apim-analytics-configs/apim-analytics-conf.yaml and set the field analyticsEnabled to "true". The other parameters also can be modified with required values.
- Create a secret with the public certificate of the wso2am-analytics server and provide the name of the created secret along with the username and password to the wso2am-analytics server (all fields must be base 64 encoded). Use the template provided for analytics-secret in apim_analytics_secret_template.yaml

##### Applying security for APIs 
- APIs created with kubernetes apim operator can be secured by defining security with security kind. It supports basic, JWT and Oauth2 security types.

   **Securing API with JWT authentication**
   
    i. Create a secret with the certificate

   `
   kubectl create secret generic <secret name> -n <namespace> --from-file=<path to cert>
   `
  
   ii. Create a security with security kind. Include the name of the secret created in step (i) in certificate field
   ```
   apiVersion: <version>
   kind: Security
   metadata:
     name: <security name>
   spec:
     type: JWT
     certificate: <name of the secret created in step 1>
     issuer: <issuer>
     audience: <audience>
   ```
   **Securing API with Oauth2 authentication**
   
    i. Create a secret with the certificate of the wso2am server
   
   `
   kubectl create secret generic <secret name> -n default --from-file=<path to cert>
   `
   
    ii. Create a secret with user credentials 
   ```
   apiVersion: v1
   kind: Secret
   metadata:
     name: <secret name>
   type: Opaque
   data:
     username: base64 encoded user name 
     password: base64 encoded password
   ```  
    iii. Create a security with security kind. Include the name of the secret created in step (i) in certificate field and name of the secret created in step (ii) in credentials field.
   ```
   apiVersion: <api_version>
   kind: Security
   metadata:
     name: <security name>
     namespace: <namespace>
   spec:
     type: Oauth
     certificate: <name of the secret created in step 1>
     endpoint: <endpoint>
     credentials: <name of the secret created in step 2>
   ```
   **NOTE:** Modify the configurations related to wso2am using the template provided in ./deploy/apim-analytics-configs/apim-analytics-conf.yaml : apim-config configmap.

   **Securing API with Basic authentication**
   
    i. Create a secret with user credentials 
   ```
   apiVersion: <version>
   kind: Secret
   metadata:
     name: <secret name>
   type: Opaque
   data:
     username: base64 encoded username 
     password: base64 encoded password
   ```
    ii. Create a security with security kind. Include the name of the secret created in step (i) in credentials field.
   ```
   apiVersion: <api_version>
   kind: Security
   metadata:
     name: <security name>
     namespace: <namespace>
   spec:
     type: basic
     credentials: <name of the secret created in step 1>
   ``` 
   **Defining the securities in swagger definition**

    Security can be defined in swagger definition under security keyword in both API and resource levels. Define the property scopes for OAuth 2 security scheme. 

   **Defining security in API level**
   
      ```
      security:
          - petstorebasic: []  
          - oauthtest: 
            - read
      ```

   **Defining security in resource level**
   
      ```
      paths:
        "/pet/findByStatus":
          get:
            security:
              - basicauth: #Oauth
                - read:pets
                - write:pets
              - petstorebasic: []
      ```


sample security definitions are provided in ./deploy/sample-definitions/security_definitions.yaml