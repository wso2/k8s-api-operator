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
    TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UZG1aak00WkRrM05qWTBZemM1TW1abU9EZ3dNVEUzTVdZd05ERTVNV1JsWkRnNE56YzRaQT09In0.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IjEwUGVyTWluIiwibmFtZSI6InNhbXBsZS1jcmQtYXBwbGljYXRpb24iLCJpZCI6NCwidXVpZCI6bnVsbH0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC93c28yYXBpbTozMjAwMVwvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoieF8xal83MW11dXZCb01SRjFLZnVLdThNOVVRYSIsImV4cCI6MzczMTQ5Mjg2MSwiaWF0IjoxNTg0MDA5MjE0LCJqdGkiOiJkYTA5Mjg2Yy03OGEzLTQ4YjgtYmFiNy1hYWZiYzhiMTUxNTQifQ.MKmGDwh855NrZ2wOvXO7TwFbCtsgsOFuoZW4DBVIbJ1KQ2F6TgTgBbtzBUvrYGPslEExMemhepfvvlYv8Gd6MMo3GVH4aO8AKyc8gHmeIQ8MQtXGn7u9N00ZW3_9JWaQkU-OYEDsLHvKKHzO0t2umaskSyCS2UkAS4wIT_szZ5sm-O-ez4nKGeJmESiV-1EchFjOhLpEH4p9wIj3MlKnZrIcJByRKK9ZGaHBqxwwYuJtMCDNa2wFAPMOh-45eabIUdo1KUO3gZLVcME93aza1t1jzL9mFsx0LGaXIxB7klrDuBCAdG9Yi3O7-3WUF74QaS2tmCxW36JhhOJ5DdacfQ
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
