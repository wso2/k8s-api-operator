### Applying security for APIs 

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
   
    i. Create a secret with the certificate of the API Portal
   
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


   sample security definitions are provided in [here](../../api-operator/deploy/sample-definitions/security_definitions.yaml)