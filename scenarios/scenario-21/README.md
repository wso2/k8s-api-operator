## Scenario 21 - Deploy a service as a managed API using WSO2 API Manager locally

- This scenario describes how to deploy a service as a managed API in a kubernetes cluster using the WSO2 API Manager deployed locally.
 

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

 ##### Steps to Follow:

- Run WSO2 API Manager locally.

- The default security custom resource needs to be deployed using the following command.
    ```
    apictl apply -f default-security.yaml
    ```
  
- Create a label in WSO2 API Manager using the admin portal giving the following values.
    ```
    Label Name: Microgateway-Internal
    Label Host: https://mgw.wso2.com:9095
    ```

- Create an API using WSO2 API Manager Publisher portal.
- Navigate to the environments tab in the publisher portal and select the label that you have created which is Microgateway-Internal. After that publish the API.
- Add the environment as dev and export the API.
    ```
    apictl add-env -e dev --apim https://localhost:9443 --token https://localhost:9443/oauth2/token
    apictl export-api -n PizzaShackAPI -v 1.0.0 -e dev -k
    ```

- Exported APIs are located under /.wso2apictl/exported/apis/dev/. Then unzip the exported folder to a desired location and deploy the API using the following command pointing to the unzipped location. 
    ```
    apictl add api -n pizzashack -f /home/jayanie/Documents/export-apis/PizzaShackAPI_1.0.0/PizzaShackAPI-1.0.0/ --override
    ```
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
    
- Get service details to invoke the API. (Please wait until the external-IP is populated in the corresponding service)
    ```
    >> apictl get services
    
    Output:
    NAME              TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                         AGE
    pizzashack        LoadBalancer   10.8.5.120     35.184.250.188   9095:32462/TCP,9090:31266/TCP   56s
    ```
    - You can see pizzashack service has been exposed as a managed API.
    - Get the external IP of the managed API's service
    
- Add an /etc/hosts entry as follows.
    ```
    <Load Balancer IP of the API>     mgw.wso2.com
    ```
- Access the API using WSO2 API Manager dev portal.
   - Create an application.
   - Subscribe to the API.
   - Generate an access token.
   - Select the Gateway environment as Microgateway-Internal.
   - Invoke the API using the try it out console.
   
 ##### How to expose an API using Ingress with WSO2 API Manager locally
 
- Run WSO2 API Manager locally.

- Create a label in WSO2 API Manager using the admin portal giving the ingress host name as
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

- Refer [Scenario-17](../scenario-17) for Prerequisites.
- Operator mode can be configured as ingress in `controller-configs/controller_conf.yaml`
     ```
      operatorMode: "ingress"
     ```
- Apply the changes
     ```$xslt
      >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
     ``` 
  
- Add the environment as dev and export the API.
     ```
      apictl add-env -e dev --apim https://localhost:9443 --token https://localhost:9443/oauth2/token
      apictl export-api -n PizzaShackAPI -v 1.0.0 -e dev -k
     ```
  
- Exported APIs are located under /.wso2apictl/exported/apis/dev/. Then unzip the exported folder to a desired location and deploy the API using the following command pointing to the unzipped location. 
     ```
      apictl add api -n pizzashack -f /home/jayanie/Documents/export-apis/PizzaShackAPI_1.0.0/PizzaShackAPI-1.0.0/ --override
     ```
      Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository
      
- Get available Ingress service
    ```
      >> apictl get ingress
      
      Output:
      NAME                              HOSTS                  ADDRESS      PORTS     AGE
      api-operator-ingress-pizzashack   mgw.ingress.wso2.com   172.17.0.2   80, 443   10m
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
