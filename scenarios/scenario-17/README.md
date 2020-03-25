## Scenario 17 - Expose API using Ingress

- This scenario showes how to expose a service using Ingress gateway.

 ***Important:***
> Follow the main README and deploy the apim-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

***Prerequrireties:***
- First install the [Nginx-ingress controller.](https://kubernetes.github.io/ingress-nginx/deploy/)
- Navigate to the apim-operator/controller-artifacts directory and set the operatorMode to "ingress" in the 
  controler_conf.yaml file.
  
  ```
  operatorMode: "ingress"
  ```
- If you have already deployed the operator you have to update operatorMode to "ingress" and apply the changes using
  following command.
  ```
  kubectl apply -f apim-operator/controller-artifacts/controler_conf.yaml
  ```
  
##### Deploying the artifacts

- Navigate to scenarios/scenario-17 directory and deploy the sample backend service using the following command.
  ```
   apictl apply -f hello-world-service.yaml
  ```
  - Output:
    ```
    targetendpoint.wso2.com/hello-world-service created
    ```
 - Basic swagger definition belongs to the "hello-world-service" service is available in swagger.yaml.
 - Create an API which is refer to the backend service "hello-world-service" using following command.
   ```
   apictl add api -n products --from-file=swagger.yaml
   ```
   - Output:
       ```
       Processing swagger 1: swagger.yaml
       creating configmap with swagger definition
       configmap/hello-world-1-swagger created
       creating API definition
       api.wso2.com/hello-world created
       ```
 - Get available API
   ```
   apictl get apis
   ```   
   - Output:
        ```
          NAME          AGE
          hello-world    1m
        ```
 - Get available Ingress service
   ```
   kubectl get ingress
   ```
   - Output:
        ```
        NAME                               HOSTS                  ADDRESS      PORTS     AGE
        api-operator-ingress-hello-world   mgw.ingress.wso2.com   34.67.56.7   80, 443   4m59s
    
        ```
    - You can see that ingress service is available for the service exposed by hello-world.
    - Using the "Host" name and IP address for the ingress service you can invoke the API.
    
 - Invoking the API 
   ```
   curl -H "Host:mgw.ingress.wso2.com" http://34.67.56.7/node/1.0.0/hello/node
   ``` 
   - Once you execute the above command, it will call to the managed API (hello-world), which then call its endpoint("hello-world-service") available in the cluster.
     If the request is success, you would be able to see the response as below.
      
     ````
     Hello World!
     ````
     
 ##### Configure SSL Passthrough in ingress 
 
 - Navigate to the apim-operator/controller-artifacts directory and set the tlsSecretName to "tls-secret" in the 
   controler_conf.yaml file.
   
 - Now you need to create a TLS certificate, using following command you can create one.
   ```
   openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=mgw.ingress.wso2.com/O=mgw.ingress.wso2.com"
   ```
   - Output:
     ```
     Generating a RSA private key
     .................................................................+++++
     ...............+++++
     writing new private key to 'tls.key'
     ```
 - Next, you need to create a tls secret using that TLS certificate.
   ```
   kubectl create secret tls tls-secret --key tls.key --cert tls.crt
   ```
   - Output:
     ```
     secret/tls-secret created
     ```
 - Now again try to invoke the API using following command, change "http" to "https" before you do the invoke. 
   ```
   curl -H "Host:mgw.ingress.wso2.com" https://34.67.56.7/node/1.0.0/hello/node -k
   ```
 - If the request is success, you would be able to see the response as below.
   ```
   Hello World!
   ```
 - Delete the API and sample backend service.
   ```
   kubectl delete api hello-world
   kubectl delete targetendpoint hello-world-service
   ```
   - Output:
     ```
     api.wso2.com "hello-world" deleted
     targetendpoint.wso2.com "hello-world-service" deleted
     ```