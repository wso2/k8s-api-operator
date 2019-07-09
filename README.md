# k8s-apim-operator

##### Navigate to the k8s-apim-operator/apim-operators/ directory and execute the following command

##### Deploy K8s client extensions

- Give executable permission to the extension files

```
chmod +x ./deploy/kubectl-extension/kubectl-add
chmod +x ./deploy/kubectl-extension/kubectl-update
chmod +x ./deploy/kubectl-extension/kubectl-show
```

- Copy the extensions to ***/usr/local/bin/***
```
cp ./deploy/kubectl-extension/kubectl-add /usr/local/bin
cp ./deploy/kubectl-extension/kubectl-update /usr/local/bin
cp ./deploy/kubectl-extension/kubectl-show /usr/local/bin
```

##### Deploy K8s CRD artifacts

- Before deploying the role you have to make yourself as a cluster admin. (Replace "email-address" with the proper value)

```
kubectl create clusterrolebinding email-address --clusterrole=cluster-admin --user=email-address
```

- Deploying CRDs for API, TargetEndpoint, Security, RateLimiting
```
kubectl apply -f ./deploy/crds/
```

- Deploying controller level configurations

"controller-configs" contains the configuration user needs to change. The docker images are created and pushed to the user's docker registry.
Update the ***user's docker registry*** in the controller_conf.yaml. Enter the base 64 encoded username and password of the user's docker registry into the docker_secret_template.yaml.

```
kubectl apply -f ./deploy/controller-configs/
```

- Deploying namespace, roles/role binding and service account associated with the operator
```
kubectl apply -f ./deploy/controller-artifacts/
```

##### Deploying an API in K8s cluster

- Deploy the API petstore
```
kubectl add api "api_name" --from-file="location to the api swagger definition"

kubectl add api petstore --from-file=./deploy/scenarios/scenario-1/petstore_basic.yaml
```
  
- Update the API petstore
```
kubectl update api "api_name" --from-file="location to the api swagger definition"

kubectl update api petstore --from-file=./deploy/scenarios/scenario-1/petstore_basic.yaml
```
  
- Delete the API
```
kubectl delete api "api_name"

kubectl delete api petstore
```

Optional Parameters

```
--replicas=3          Number of replicas
--namespace=wso2      Namespace to deploy the API

kubectl add api "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" --namespace="desired namespace"
```

***Note:***
Namespace and replicas are optional parameters. If they are not provided default namespace will be used and 1 replica will be created. However, the namespace used in all the commands related to particular API name must match.

##### Cleanup

```
kubectl delete -f ./deploy/controller-artifacts/
kubectl delete -f ./deploy/controller-configs/
kubectl delete -f ./deploy/crds/
```
    
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
   kubectl create secret generic <secret name> -n <namespace> --from-file=<path to cert>
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
   apiVersion: <version>
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
   apiVersion: v1
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
   apiVersion: <version>
   kind: Security
   metadata:
     name: <security name>
     namespace: <namespace>
   spec:
     type: basic
     credentials: <name of the secret created in step 1>
   ``` 
   **Defining the securities in swagger definition**

    Security can be defined in swagger definition under security keyword in both API and resource levels. Define the property scopes for OAuth2 security scheme. 

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
              - basicauth:
                - read:pets
                - write:pets
              - petstorebasic: []
     ```


   sample security definitions are provided in ./deploy/sample-definitions/security_definitions.yaml