### Applying security for APIs 

- APIs created with kubernetes apim operator can be secured by defining security with security kind. It supports basic, JWT and Oauth2 security types.
- Note: When a security custom resource refers a secret, you need to make sure the namespace of the secret is same as the namespace of that security custom resource.
        When an API refers a security custom resource in swagger definition under security keyword you need to make sure that the namespace of the security custom 
        resource is same as the namespace that the API belongs to. 
#### Securing API with JWT authentication
   
1. Create a secret with the certificate

   ```sh
   >> kubectl create secret generic <SECRET_NAME> -n <NAMESPACE> --from-file=<PATH_TO_CERT>
   ```
  
  - The namespace of the secret should be the namespace of the security custom resource.  
1. Create a security with security kind. Include the name of the secret created in step (1) in certificate field
   ```yaml
   apiVersion: <VERSION>
   kind: Security
   metadata:
     name: <SECURITY_NAME>
     namespace: <NAMESPACE>
   spec:
     type: JWT
     securityConfig:
       - issuer: <ISSUER>
         audience:  <AUDIENCE>
         certificate: <NAME_OF_THE_SECRET_CREATED_IN_STEP_1>
   ```
#### Securing API with Oauth2 authentication

1. Create a secret with the certificate
   ```sh
   >> kubectl create secret generic <SECRET_NAME> -n <NAMESPACE> --from-file=<PATH_TO_CERT>
   ```
   - The namespace of the secret should be the namespace of the security custom resource.
1. Create a secret with user credentials 
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: <SECRET_NAME>
   type: Opaque
   data:
     username: <BASE64_ENCODED_USER_NAME>
     password: <BASE64_ENCODED_PASSWORD>
   ```  
   
1. Create a security with security kind. Include the name of the secret created in step (1) in certificate field and name of the secret created in step (2) in credentials field.
   ```yaml
   apiVersion: <VERSION>
   kind: Security
   metadata:
     name: <SECURITY_NAME>
     namespace: <NAMESPACE>
   spec:
     type: Oauth
     securityConfig:
       - certificate: <NAME_OF_THE_SECRET_CREATED_IN_STEP_1>
         endpoint: <ENDPOINT>
         credentials: <NAME_OF_THE_SECRET_CREATED_IN_STEP_2>
   ```
   
#### Securing API with Basic authentication
   
1. Create a secret with user credentials 
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: <SECRET_NAME>
   type: Opaque
   data:
     username: <BASE64_ENCODED_USER_NAME>
     password: <BASE64_ENCODED_PASSWORD>
   ```
   
1. Create a security with security kind. Include the name of the secret created in step (1) in credentials field.
   ```yaml
   apiVersion: <version>
   kind: Security
   metadata:
     name: <SECURITY_NAME>
     namespace: <NAMESPACE>
   spec:
     type: basic
     securityConfig:
       - credentials: <NAME_OF_THE_SECRET_CREATED_IN_STEP_1>
   ``` 

#### Defining the securities in swagger definition

Security can be defined in swagger definition under security keyword in both API and resource levels. Define the property scopes for OAuth2 security scheme. 

1. Defining security in API level**
   
     ```yaml
      security:
          - petstorebasic: []
          - oauthtest:
            - read
     ```

1. Defining security in resource level**
   
     ```yaml
      paths:
        "/pet/findByStatus":
          get:
            security:
              - basicauth:
                - read:pets
                - write:pets
              - petstorebasic: []
     ```


Sample security definitions are provided in [here](../../../api-operator/deploy/sample-definitions/security_definitions.yaml)