### Working with endpoints 

- During the deployment of APIs created with the apim operator sometimes it requires to deploy the endpoint services
  associated with those APIs. The target endpoint kind provides the flexibility to deploy the backend services by specifying the 
  relevant docker images and parameters. 

   #### Deploy endpoints using the target endpoint kind
   
    ##### i. Create a endpoint with target endpoint kind
    
    The template for the target endpoint kind can be specified as follow
    
   ```
   apiVersion: <version>
   kind: TargetEndpoint
   metadata:
     name: <endpoint service name>
     namespace: wso2-system
     labels:
       app: wso2
   spec:
     protocol: <https or http>
     port: <port>
     targetPort: <target port>
     deploy:
       name: <endpoint name>
       dockerImage: <docker-image>
       count: <replica count>
     mode : <privateJet or sidecar>
   ```
   Sample definition of endpoint kind can be specified as follow
   
    ```
   apiVersion: wso2.com/v1alpha1
   kind: TargetEndpoint
   metadata:
     name: simple-endpoint-service
     namespace: wso2-system
     labels:
       app: wso2
   spec:
     protocol: https
     targetPort: 9443
     port: 443
     deploy:
       name: simple-endpoint
       dockerImage: pubudu/review:1.0.0
       count: 3
     mode : privateJet
    ```
    ##### ii. Deploy the target endpoint
    The target endpoint definition can be deployed using the command line tool as follow
    
    ```
        apictl apply -f sample-endpoint-service.yaml
    ```
    - Output:
    ```
        targetendpoint.wso2.com/sample-endpoint-service created
    ```        
    
    ##### iii. Target endpoint parameters
    
    Target endpoint template which specified above contains several parameters which used to define specific functions.
    Here are the usages of the important parameters in the definition
    
     ```
       ...
       metadata:
         name: <endpoint service name>  // When endpoint is deployed, it will be associated with a service. 
                                        // The name specified here will be used as the service name. This name also will be used to discover the target endpoint
         namespace: wso2-system
         labels:
           app: wso2                    // The name space and label shouldn't be changed 
       spec:
         protocol: <https or http>      // Specify the protocol that service should be exposed
         port: <port>                   // If the port and target port do not specified, depend on the protocol type
         targetPort: <target port>      // port and target port will be assigned. https associated with port 443 and target port 443 and
                                        // http associated with port 80 and target port 80     
           deploy:                       
           name: <endpoint name>        // name of the endpoint 
           dockerImage: <docker-image>  // docker image name
           count: <replica count>       // number of replicas that should be deployed
         mode : <privateJet or sidecar> // Mode is very important paramets in the target endpoint kind. If the mode is set to privateJet
                                        // target endpoint controller will create the endpoint deployment along with the service.
                                        // In sidecar mode, target endpoint controller only add the endpoint definition but no deployment will be created 
     ```
    
    ##### iv. How target endpoint works
    
    Using apim operator provide greater flexibility to deploy APIs in Kurbenetes. Target endpoint can be deployed in two modes. Target endpoint
    default mode is set to privateJet which means target endpoint controller will create the deployment and service for the endpoint using the
    docker image specified in the endpoint definition. 
    
    When mode is set to sidecar, target endpoint controller will only add the definition to kubernetes registry and will not create the deployment
    as operated in the privatejet mode. In the sidecar mode, user can specify the name of the endpoint service and mode in swagger definition as follow.
      
    Eg :<br>
    
    ```
    x-wso2-production-endpoints: simple-endpoint-service 
    
    x-wso2-mode: sidecar   
    ```    
    At the time of the API is deployed, apim-operator will identify the endpoint service and create the target endpoint deployment along with the
    service. The deployed API will act as a sidecar to the deployed endpoint.

   - Sample security definitions are provided in `<k8s-api-operator-home>/apim-operator/deploy/sample-definitions/wso2_v1alpha1_targetendpoint_cr.yaml`
   - Sample scenarios using the target endpoint provided in `<k8s-api-operator-home>/scenarios/scenario-7` and `<k8s-api-operator-home>/scenarios/scenario-8`