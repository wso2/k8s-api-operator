## Scenario 17 - Expose an API using Ingress

- This scenario shows how to expose a service using the Ingress Controller.

**Important:**
> Follow the main README and deploy the api-operator and configuration files.

**Prerequisites:**
- First install the [Nginx-ingress controller.](https://kubernetes.github.io/ingress-nginx/deploy/)
- Navigate to the api-operator/controller-artifacts directory and set the operatorMode to "ingress" in the 
  controler_conf.yaml file.
  
  ```
  operatorMode: "ingress"
  ```
- If you have already deployed the operator you have to update operatorMode to "ingress" and apply the changes using
  following command.
  ```
  >> apictl apply -f api-operator/controller-artifacts/controller_conf.yaml
  ```
  
#### Deploying the artifacts

- Navigate to scenarios/scenario-17 directory and deploy the sample backend service using the following command.
    ```
    >> apictl apply -f hello-world-service.yaml
    
    Output:
    targetendpoint.wso2.com/hello-world-service created
    ```
    Basic swagger definition belongs to the "hello-world-service" service is available in swagger.yaml.

- Create an API which is refer to the backend service "hello-world-service" using following command.
    
    ```
    >> apictl add api -n hello-world-api --from-file=swagger.yaml
    
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/hello-world-1-swagger created
    creating API definition
    api.wso2.com/hello-world created
    ```

 - Get available Ingress service
 
    ```
    >> apictl get ingress
    
    Output:
    NAME                               HOSTS                  ADDRESS      PORTS     AGE
    api-operator-ingress-hello-world   mgw.ingress.wso2.com   34.67.56.7   80, 443   4m59s
    ```
    - You can see that ingress service is available for the service exposed by hello-world.
    - Using the "Host" name and IP address for the ingress service you can invoke the API.
    
 - Invoking the API 
 
    ```
    TOKEN=eyJ4NXQiOiJNell4TW1Ga09HWXdNV0kwWldObU5EY3hOR1l3WW1NNFpUQTNNV0kyTkRBelpHUXpOR00wWkdSbE5qSmtPREZrWkRSaU9URmtNV0ZoTXpVMlpHVmxOZyIsImtpZCI6Ik16WXhNbUZrT0dZd01XSTBaV05tTkRjeE5HWXdZbU00WlRBM01XSTJOREF6WkdRek5HTTBaR1JsTmpKa09ERmtaRFJpT1RGa01XRmhNelUyWkdWbE5nX1JTMjU2IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhdWQiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwibmJmIjoxNTk3MjExOTUzLCJhenAiOiJKRmZuY0djbzRodGNYX0xkOEdIVzBBR1V1ME1hIiwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL3dzbzJhcGltOjMyMDAxXC9vYXV0aDJcL3Rva2VuIiwiZXhwIjoxOTMwNTQ1Mjg2LCJpYXQiOjE1OTcyMTE5NTMsImp0aSI6IjMwNmI5NzAwLWYxZjctNDFkOC1hMTg2LTIwOGIxNmY4NjZiNiJ9.UIx-l_ocQmkmmP6y9hZiwd1Je4M3TH9B8cIFFNuWGHkajLTRdV3Rjrw9J_DqKcQhQUPZ4DukME41WgjDe5L6veo6Bj4dolJkrf2Xx_jHXUO_R4dRX-K39rtk5xgdz2kmAG118-A-tcjLk7uVOtaDKPWnX7VPVu1MUlk-Ssd-RomSwEdm_yKZ8z0Yc2VuhZa0efU0otMsNrk5L0qg8XFwkXXcLnImzc0nRXimmzf0ybAuf1GLJZyou3UUTHdTNVAIKZEFGMxw3elBkGcyRswzBRxm1BrIaU9Z8wzeEv4QZKrC5NpOpoNJPWx9IgmKdK2b3kIWJEFreT3qyoGSBrM49Q
    ```
    
    ```
    >> curl -H "Host:mgw.ingress.wso2.com" https://34.67.56.7/node/1.0.0/hello/node -H "Authorization:Bearer $TOKEN" -k
    
    Output:
    Hello World!
    ```
   
   The default configuration uses the Passthrough approach in Ingress.
   
   ```
   [Client] --- http/https —> [Ingress] --- https -—> [API Microgateway]
   ```
<br>

#### How to expose an API using a different Ingress Hostname

- When you are creating the API you can define a hostname if required. If a hostname is not defined
it takes the default ingress hostname (mgw.ingress.wso2.com).

   ```
        >> apictl add api -n hello-world-api --from-file=swagger.yaml --override --hostname=mgw.group1.wso2.com
    
    Output:
    Processing swagger 1: swagger.yaml
    creating configmap with swagger definition
    configmap/hello-world-1-swagger created
    creating API definition
    api.wso2.com/hello-world created
   ```
   Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository

- Get available Ingress service
   
   ```
      >> apictl get ingress
      
      Output:
      NAME                               HOSTS                  ADDRESS      PORTS     AGE
      api-operator-ingress-hello-world   mgw.group1.wso2.com   34.67.56.7   80, 443   4m59s
   ```

#### Configure No TLS in ingress

In the no tls approach, it has the following flow and it uses only http protocol in the request flow.

```
[Client] --- http —> [Ingress] --- http -—> [API Microgateway]
```

- Change the followings in the ingress-config configmap located in api-operator/controller/artifacts/.

    - Set **ingressTransportMode** to http
    - Set **nginx.ingress.kubernetes.io/backend-protocol** annotation to HTTP
    
     ```
     apiVersion: v1
     kind: ConfigMap
     metadata:
       name: ingress-configs
       namespace: wso2-system
     data:
       ingress.properties: |
         nginx.ingress.kubernetes.io/backend-protocol: HTTP
         kubernetes.io/ingress.class: nginx
         nginx.ingress.kubernetes.io/ssl-redirect: false
       ingressResourceName: "api-operator-ingress"
       ingressTransportMode: "http"
       ingressHostName : "mgw.ingress.wso2.com"
       #tlsSecretName: ""
     ```

#### Configure Edge TLS in ingress

In the approach, it has the following flow and it uses only https protocol between client and ingress controller.

```
[Client] --- https —> [Ingress] --- http -—> [API Microgateway]
```

- Change the followings in the ingress-config configmap located in api-operator/controller/artifacts/.

    - Set **ingressTransportMode** to http
    - Set **nginx.ingress.kubernetes.io/backend-protocol** annotation to HTTP
    
     ```
     apiVersion: v1
     kind: ConfigMap
     metadata:
       name: ingress-configs
       namespace: wso2-system
     data:
       ingress.properties: |
         nginx.ingress.kubernetes.io/backend-protocol: HTTP
         kubernetes.io/ingress.class: nginx
         nginx.ingress.kubernetes.io/ssl-redirect: false
       ingressResourceName: "api-operator-ingress"
       ingressTransportMode: "http"
       ingressHostName : "mgw.ingress.wso2.com"
       #tlsSecretName: ""
     ```
  
#### Configure Re-encrypt in ingress 

In the approach, it has the following flow and it uses only https protocol in the request flow. In the ingress controller it re-encrypts. 

```
[Client] --- https —> [Ingress] --- https -—> [API Microgateway]
```

 - Navigate to the api-operator/controller-artifacts directory and set the tlsSecretName to "tls-secret" in the 
   controler_conf.yaml file.
   
 - Create a TLS certificate, using following command.
 
    ```
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=mgw.ingress.wso2.com/O=mgw.ingress.wso2.com"
    
    Output:
    Generating a RSA private key
    .................................................................+++++
    ...............+++++
    writing new private key to 'tls.key'
    ```
   
 - Create a tls secret using that TLS certificate.
 
    ```
    >> apictl create secret tls tls-secret --key tls.key --cert tls.crt -n wso2-system
    
    Output:
    secret/tls-secret created
    ```
   
- Change the followings in the ingress-config configmap located in api-operator/controller/artifacts/.

    - Set **ingressTransportMode** to https
    - Set **nginx.ingress.kubernetes.io/backend-protocol** annotation to HTTPS
    - Set **tlsSecretName** to tls-secret
    
     ```
     apiVersion: v1
     kind: ConfigMap
     metadata:
       name: ingress-configs
       namespace: wso2-system
     data:
       ingress.properties: |
         nginx.ingress.kubernetes.io/backend-protocol: HTTPS
         kubernetes.io/ingress.class: nginx
         nginx.ingress.kubernetes.io/ssl-redirect: false
       ingressResourceName: "api-operator-ingress"
       ingressTransportMode: "https"
       ingressHostName : "mgw.ingress.wso2.com"
       tlsSecretName: "tls-secret"
     ```
