## Scenario 21 - Deploy a service as a managed API using WSO2 API Manager locally

- This scenario describes how to deploy a service as a managed API in a Kubernetes cluster using the WSO2 API Manager deployed locally.
 
 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

The following options are available.

1. Publish an API in API Manager and deploy an API gateway using the Ingress mode
2. Publish an API in API Manager and deploy an API gateway using Load Balancer mode (default mode)
3. Deploy an API gateway in Ingress/Load Balancer mode and publish an API in API Manager


#### Publish an API in API Manager and deploy an API gateway using the Ingress mode
 
- Run WSO2 API Manager locally.

- Create a label in WSO2 API Manager using the admin portal (https://localhost:9443/admin) giving the ingress host name as
the label host.

    ```
     Label Name: Microgateway-Ingress
     Label Host: https://mgw.ingress.wso2.com
    ```
  
- Create an API using WSO2 API Manager Publisher portal.

- Navigate to the environments tab in the publisher portal and select the label that you have created which is 
Microgateway-Ingress. After that publish the API.

- The default security custom resource needs to be deployed using the following command.

     ```
      apictl apply -f default-security.yaml
     ```

- Install the [Nginx-ingress controller.](https://kubernetes.github.io/ingress-nginx/deploy/)

- Operator mode can be configured as ingress in `<k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml`
     ```
      operatorMode: "ingress"
     ```
- Apply the changes
     ```$xslt
      >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
     ``` 
  
- Create an API using the API Publisher. When publishing the API, select Microgateway-Ingress Gateway as the environment.


- Add the environment as dev and export the API.

     ```
      >> apictl add-env -e dev --apim https://localhost:9443 --token https://localhost:9443/oauth2/token
      >> apictl export-api -n HelloWorldAPI -v 1.0.0 -e dev -k
     ```
  
- Exported APIs are located under /.wso2apictl/exported/apis/dev/. Then unzip the exported folder to a desired location and deploy the API using the following command pointing to the unzipped location. 

     ```
      >> apictl add api -n helloworld-api -f ./HelloWorldAPI-1.0.0/ --override
     ```
  
  Note: **--override** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
      
- Get available Ingress service

    ```
      >> apictl get ingress
      
      Output:
      NAME                                  HOSTS                  ADDRESS      PORTS     AGE
      api-operator-ingress-helloworld-api   mgw.ingress.wso2.com   172.17.0.2   80, 443   10m
    ```
     - You can see that ingress service is available for the service exposed by pizzashack.
     - Using the "Host" name and IP address for the ingress service you can invoke the API.
      
- Add an /etc/hosts entry as follows.

    ```
     <IP Address of Ingress service>     <Host name of Ingress service>
    ```
  
- Access the API using WSO2 API Manager dev portal.
    - Create an application.
    - Subscribe to the API.
    - Generate an access token.
    - Select the Gateway environment as Microgateway-Ingress.
    - Invoke the API using the try it out console.


#### Publish an API in API Manager and deploy an API gateway using Load Balancer mode (default mode)

- Run WSO2 API Manager locally.

- The default security custom resource needs to be deployed using the following command.

    ```
    apictl apply -f default-security.yaml
    ```

- Create and publish an API using WSO2 API Manager Publisher portal.

- Add the environment as dev and export the API.
    ```
    >> apictl add-env -e dev --apim https://localhost:9443 --token https://localhost:9443/oauth2/token
    >> apictl export-api -n HelloWorldAPI -v 1.0.0 -e dev -k
    ```

- Exported APIs are located under /.wso2apictl/exported/apis/dev/. Then unzip the exported folder to a desired location and deploy the API using the following command pointing to the unzipped location. 
    ```
    >> apictl add api -n helloworld-api -f ./HelloWorldAPI --override
    ```
  
  Note: **--override** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
    
- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME              TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    helloworld-api    LoadBalancer   10.8.5.120     35.184.250.188   9095:32462/TCP,9090:31266/TCP   56s
    ```
    - You can see helloworld-api service has been exposed as a managed API.
    - Get the external IP of the managed API's service
    
- Access the API using WSO2 API Manager dev portal.
   - Create an application.
   - Subscribe to the API.
   - Generate an access token.
   - Access the API using the curl command. (In the Load Balancer type, CORs are not handled. Hence it cannot access via the try it out console.)

    ```
    >> curl -X GET "https://<EXTERNAL_IP>:9095/<API-context>/<API-resource>"  -H "Authorization:Bearer $TOKEN" -k
    ```
   
#### Deploy an API gateway in Ingress/Load Balancer mode and publish an API in API Manager

- The default security custom resource needs to be deployed using the following command.

     ```
      apictl apply -f default-security.yaml
     ```

- Install the [Nginx-ingress controller.](https://kubernetes.github.io/ingress-nginx/deploy/)

- Operator mode can be configured as ingress in `<k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml`
     ```
      operatorMode: "ingress"
     ```
  
- Apply the changes
     ```$xslt
      >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
     ``` 

- Deploy the API

     ```
      >> apictl add api -n petstore-api --from-file=swagger.yaml --override
     ```
  
  Note: **--override** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.
      
- Get available Ingress service

    ```
      >> apictl get ingress
      
      Output:
      NAME                               HOSTS                  ADDRESS      PORTS     AGE
      api-operator-ingress-petstore-api  mgw.ingress.wso2.com   172.17.0.2   80, 443   10m
    ```
     - You can see that ingress service is available for the service exposed by petstore-api.
     - Using the "Host" name and IP address for the ingress service you can invoke the API.
      
- Add an /etc/hosts entry as follows.

    ```
     <IP Address of Ingress service>     <Host name of Ingress service>
    ```
    
- Run WSO2 API Manager locally.

- Create a label in WSO2 API Manager using the admin portal (https://localhost:9443/admin) giving the ingress host name as
the label host.

    ```
     Label Name: Microgateway-Ingress
     Label Host: https://mgw.ingress.wso2.com
    ```

- Initialize a API project with the swagger definition.
    ```sh
    >> apictl init petstore-api \
              --oas=./swagger.yaml \
              --initial-state=PUBLISHED
    ```

- Edit the `petstore-api/Meta-information/api.yaml` with adding `gatewayLabels` as follows.
    ```yaml
    gatewayLabels:
      - name: Microgateway-Ingress
    ```

- Import the API to API Manager
    ```sh
    >> apictl add-env \
            -e dev \
            --apim https://localhost:9443 \
            --token https://localhost:9443/oauth2/token
  
    >> apictl import-api -f petstore-api/ -e dev -k 
    ```  

- Access the API using WSO2 API Manager dev portal.
    - Create an application.
    - Subscribe to the API.
    - Generate an access token.
    - Select the Gateway environment as Microgateway-Ingress.
    - Invoke the API using the try it out console.

