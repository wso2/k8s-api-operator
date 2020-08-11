### Working with endpoints 

- During the deployment of APIs created with the apim operator sometimes it requires to deploy the endpoint services
  associated with those APIs. The target endpoint kind provides the flexibility to deploy the backend services by specifying the 
  relevant docker images and parameters. 

#### Deploy endpoints using the target endpoint kind
   
1. Create a endpoint with target endpoint kind
   
   The template for the target endpoint kind can be specified as follow
    
   ```yaml
   apiVersion: <VERSION>
   kind: TargetEndpoint
   metadata:
     name: <ENDPOINT_SERVICE_NAME>
     namespace: <NAMESPACE>
     labels:
       app: wso2
   spec:
     applicationProtocol: <https_OR_http>
     ports:
       - name: <PORT_NAME>
         port: <PORT>
         targetPort: <TARGET_PORT>
     mode: <privateJet_OR_sidecar>
     deploy:
       name: <DEPLOYMENT_NAME>
       dockerImage: <DOCKER_IMAGE>
       minReplicas: <MINIMUM_REPLICA_COUNT>
       maxReplicas: <MAXIMUM_REPLICA_COUNT>
       requestCPU: <REQUEST_CPU>
       reqMemory: <REQUEST_MEMORY>
       cpuLimit: <CPU_LIMIT>
       memoryLimit: <MEMORY_LIMIT>
   ```
   
   Sample definition of endpoint kind can be specified as follow
   
    ```yaml
    apiVersion: wso2.com/v1alpha1
    kind: TargetEndpoint
    metadata:
      name: products-privatejet
      labels:
        app: wso2
    spec:
      applicationProtocol: http
      ports:
        - name: prod-ep
          port: 80
          targetPort: 9090
      deploy:
        name: products-pj-service
        dockerImage: pubudu/products:1.0.0
        minReplicas: 2
        maxReplicas: 3
        requestCPU: "60m"
        reqMemory: "32Mi"
        cpuLimit: "120m"
      mode: privateJet
    ```

1. Deploy the target endpoint
    The target endpoint definition can be deployed using the command line tool as follow
    
    ```sh
    >> apictl apply -f sample-endpoint-service.yaml
   
    - Output:
    targetendpoint.wso2.com/sample-endpoint-service created
    ```        
    
1. Target endpoint parameters
    
    Target endpoint template which specified above contains several parameters which used to define specific functions.
    Here are the usages of the important parameters in the definition
    
   ```yaml
   spec:
     applicationProtocol: <https_OR_http>    // Specify the protocol that service should be exposed
     ports:                                  // Ports of the target endpoint
       - name: <PORT_NAME>                   // Name of the port
         port: <PORT>                        // The port that will be exposed by this service
         targetPort: <TARGET_PORT>           // Port that is targeted to expose
     mode: <privateJet_OR_sidecar>           // Mode is very important paramets in the target endpoint kind.
                                             // If the mode is set to privateJet, Target endpoint controller will create
                                             // the endpoint deployment along with the service.
                                             // In sidecar mode, target endpoint controller only add the endpoint
                                             // definition but no deployment will be created
     deploy:
       name: <DEPLOYMENT_NAME>               // Name of the deployment
       dockerImage: <DOCKER_IMAGE>           // Docker image to deploy
       minReplicas: <MINIMUM_REPLICA_COUNT>  // Number of minimum replicas that should be deployed
       maxReplicas: <MAXIMUM_REPLICA_COUNT>  // Number of maximum replicas that should be deployed
       requestCPU: <REQUEST_CPU>             // Minimum CPU required to deploy the pod
       reqMemory: <REQUEST_MEMORY>           // Minimum memory required to deploy the pod
       cpuLimit: <CPU_LIMIT>                 // Maximum CPU value for the pod can survive
       memoryLimit: <MEMORY_LIMIT>           // Maximum memory value for the pod can survive
   ```
    
1. How target endpoint works
    
    API Operator provides greater flexibility to deploy APIs in Kubernetes. Target endpoint can be deployed in two
    modes. Target endpoint's default mode is set to **privateJet** which means target endpoint controller will create
    the deployment and service for the endpoint using the specified docker image in the endpoint definition. 
    
    When mode is set to **sidecar**, target endpoint controller will only add the definition to kubernetes registry
    and will not create the deployment as operated in the privatejet mode. In the sidecar mode, user can specify
    the name of the endpoint service and mode in swagger definition as follow.
      
    Eg :<br>
    
    ```yaml
    x-wso2-production-endpoints:
      urls:
        - simple-endpoint-service
    x-wso2-mode: sidecar
    ```    
    At the time of the API is deployed, API Operator will identify the endpoint service and create the target endpoint
    deployment along with the service. The deployed API will act as a sidecar to the deployed endpoint.
    
    - Sample target endpoint definitions are provided in `<k8s-api-operator-home>/api-operator/deploy/sample-definitions/wso2_v1alpha1_targetendpoint_cr.yaml`
    - Sample scenarios using the target endpoint provided in `<k8s-api-operator-home>/scenarios/scenario-7` and
      `<k8s-api-operator-home>/scenarios/scenario-8`