### Configuration Overview

#### How to configure Readiness and Liveness probes

- Readiness and Liveness probes are responsible to check the health status of your deployment and pods.

- Readiness probe describe if the pod is ready to accept the traffic.

- Liveness probe describe the health status of the pod.

- Depending on you environment, you might want to change these values.

- Open \<k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml

- Following are the default values present in the configuration file.

- Depending on your environment, you may change the values.
    ```yaml
      #Configure readiness probe initial delay for API pod
      readinessProbeInitialDelaySeconds: "8"
      #Configure readiness prob interval for API pod
      readinessProbePeriodSeconds: "5"
      #Configure liveness probe initial delay for API pod
      livenessProbeInitialDelaySeconds: "10"
      #Configure liveness probe interval for API pod
      livenessProbePeriodSeconds: "30"
    ```
- Once you done these changed you have to execute the following command to apply these changed in the Kubernetes cluster.
    ```sh
    >> kubectl apply -f api-operator/controller-configs/controller_conf.yaml
    ```
  
#### How to configure Microgateway

- Default configurations

    Default configurations related to Microgateway can be found in \<k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml.
    You can change these configurations depending on the availability of resources.
    
    ```yaml
    #mgw toolkit image to initialize/setup the micro gw project
    mgwToolkitImg: wso2am/wso2micro-gw-toolkit:3.2.0
    #mgw runtime image to use in the mgw executable
    mgwRuntimeImg: wso2/wso2micro-gw:3.2.0
    #kaniko image for the kaniko pod which builds the mgw api runtime and pushes to the registry
    kanikoImg: gcr.io/kaniko-project/executor:v0.24.0
    #Required CPU usage for pods.   Default-> resourceRequestCPU: "1000m"
    resourceRequestCPU: "1000m"
    #Required Memory usage pods can use.   Default->  resourceRequestMemory: "512Mi"
    resourceRequestMemory: "512Mi"
    #Max CPU usage limit a pod can use.   Default->  resourceLimitCPU: "2000m"
    resourceLimitCPU: "2000m"
    #Max Memory usage limit a pod can use.   Default->  resourceLimitMemory: "512Mi"
    resourceLimitMemory: "512Mi"
    
    resourceRequestCPUTarget: "500m"
    #Required Memory usage pods can use for TargetEndPoint.   Default->  resourceRequestMemory: "512Mi"
    resourceRequestMemoryTarget: "512Mi"
    #Max CPU usage limit a pod can use for TargetEndPoint.   Default->  resourceLimitCPU: "2000m"
    resourceLimitCPUTarget: "500m"
    #Max Memory usage limit a pod can use for TargetEndPoint.   Default->  resourceLimitMemory: "512Mi"
    resourceLimitMemoryTarget: "512Mi"
    #Configure readiness probe initial delay for API pod
    readinessProbeInitialDelaySeconds: "8"
    #Configure readiness prob interval for API pod
    readinessProbePeriodSeconds: "5"
    #Configure liveness probe initial delay for API pod
    livenessProbeInitialDelaySeconds: "10"
    #Configure liveness probe interval for API pod
    livenessProbePeriodSeconds: "30"
    #Stop at docker image creation or continue to deploy kubernetes artifact.
    #Default->  generatekubernbetesartifactsformgw: "true"
    generatekubernbetesartifactsformgw: "true"
    #Available modes - ingress, default , route and clusterIP
    operatorMode: "default"
    #Expose custom metrics. Default-> observabilityEnabled: "false"
    observabilityEnabled: "false"
    ``` 
    
    ```yaml
    #By default hostname verification is disabled. In a production scenario, this has to be enabled.
    verifyHostname: "false"
    #Log level of the managed API (microgateway). Available levels: INFO, DEBUG, TRACE
    logLevel: "INFO"
    #Ports from which the managed API service is getting exposed
    httpPort: "9090"
    httpsPort: "9095"
    #Enable distributed ratelimiting. Default value:false. If enabled please deploy API Portal
    enabledGlobalTMEventPublishing: "false"
    #The central traffic management solution URL (related to distributed ratelimiting)
    #Format: hostname_of_API_Portal:Default_port
    throttleEndpoint: "wso2apim.wso2:9443"
    #Message broker connection URL (related to distributed ratelimiting and token revocation)
    #Format: hostname_of_API_Portal:JMS_port
    jmsConnectionProvider: "wso2apim.wso2:5672"
    #Token revocation
    #Enable real time notifier for token revocation
    enableRealtimeMessageRetrieval: "false"
    #Request and response validation
    enableRequestValidation: "false"
    enableResponseValidation: "false"
    #APIKey issuer configurations
    #APIKey STS token configurations
    enabledAPIKeyIssuer: "true"
    apiKeyKeystorePath: "${mgw-runtime.home}/runtime/bre/security/ballerinaKeystore.p12"
    apiKeyKeystorePassword: "ballerina"
    apiKeyIssuerName: "https://localhost:9095/apikey"
    apiKeyIssuerCertificateAlias: "ballerina"
    validityTime: "-1"
    #Provide the list of allowed APIs by the generated API Key
    allowedAPIs: |
    # - API name given in the API Definition: Allowed versions
    ```

- ##### Ingress Mode

    To use the Ingress controller, change the operator mode to "ingress". This can be found under default
    configurations for Microgateway.
    
    Ingress specific configurations can also be changed in controller_conf.yaml.
    
    ```yaml
    ingress.properties: |
        nginx.ingress.kubernetes.io/backend-protocol: HTTPS
        kubernetes.io/ingress.class: nginx
        nginx.ingress.kubernetes.io/ssl-redirect: false
        nginx.ingress.kubernetes.io/enable-cors: true
        nginx.ingress.kubernetes.io/cors-allow-origin: *
        nginx.ingress.kubernetes.io/cors-allow-methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
        nginx.ingress.kubernetes.io/cors-allow-headers: authorization, Access-Control-Allow-Origin, Content-Type, SOAPAction, apikey, Authorization
    ingressResourceName: "api-operator-ingress"
    #Define whether ingress to use http or https endpoint of operator deployment
    ingressTransportMode: "https"
    #Define the hostname of the ingress
    ingressHostName : "mgw.ingress.wso2.com"
    #Define the secret name for TLS certificate
    #tlsSecretName: ""
    ```
  
- ##### Route Mode
    
    To expose an API using Openshift Route, change the operator mode to "route". This can be found under default
    configurations for Microgateway.
    
    Route specific configurations can also be changed in controller_conf.yaml
    
    ```yaml
    route.properties: |
      openshift.io/host.generated: false
    routeName: "api-operator-route"
    #Define whether Route to use http or https endpoint of operator deployment
    routeTransportMode: "http"
    #Define the hostname of the Route
    routeHost : "mgw.route.wso2.com"
    # TLS termination - passthrough, edge, reencrypt
    tlsTermination: ""
    ``` 

- ##### Istio Mode
    
    For applying API management for microservices that are deployed in Istio, change the operator mode to "istio". 
    This can be found under default configurations for Microgateway.
    
    Istio specific configurations can also be changed in controller_conf.yaml
    
    ```yaml
    #Gateway name
    gatewayName: "wso2-gateway"
    #Gateway host
    host: "internal2.wso2.com"
    #CORS policy
    corsPolicy: |
    allowOrigins:
      - exact: https://localhost:9443
    allowMethods:
      - GET
      - POST
      - PUT
      - DELETE
    allowCredentials: true
    allowHeaders:
      - authorization
      - Access-Control-Allow-Origin
      - Content-Type
      - SOAPAction
      - apikey
      - Authorization
    ``` 

Once you have done any changes to above configs, you have to execute the following command to apply changes to the cluster.

```shell script
>> kubectl apply -f api-operator/controller-configs/controller_conf.yaml
``` 

- ##### Advanced Configurations

    You can further change the configurations related to Microgateway by changing the
    /<k8s-api-operator-home>/api-operator/controller-configs/mgw_conf_mustache.yaml file.
    
    Make sure to execute following command for your changes to take effect in the cluster.
    ```shell script
    >> kubectl apply -f api-operator/controller-configs/mgw_conf_mustache.yaml
    ```


  
#### How to change HPA(Horizontal Pod Autoscaler) related configurations

- API Operator provides the HPA capability to the deployed API.
- HPA will be populated from the default values.
- These configurations reside in the \<k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml
    - Find the default values below.
      ```yaml
      # Horizontal Pod Auto-Scaling for Micro-Gateways
      # Maximum number of replicas for the Horizontal Pod Auto-scale. Default->  maxReplicas: "5"
      mgwMaxReplicas: "5"
      # Metrics configurations for v2beta2
      mgwMetrics: |
        - type: Resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 50
        # - type: Pods
        #   pods:
        #     metric:
        #       name: http_requests_total_value_per_second
        #     target:
        #       type: AverageValue
        #       averageValue: 100m
        # - type: Object
        #   object:
        #     metric:
        #       name: requests-per-second
        #     describedObject:
        #       apiVersion: networking.k8s.io/v1beta1
        #       kind: Ingress
        #       name: main-route
        #     target:
        #       type: Value
        #       value: 10k
    
      # Metrics Configurations for v2beta1
      mgwMetricsV2beta1: |
        - type: Resource
          resource:
            name: cpu
            targetAverageUtilization: 50
    
      # Horizontal Pod Auto-Scaling for Target-Endpoints
      # Maximum number of replicas for the Horizontal Pod Auto-scale. Default->  maxReplicas: "5"
      targetEndpointMaxReplicas: "5"
      # Metrics configurations for v2beta2
      targetEndpointMetrics: |
        - type: Resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 50
    
      # Metrics Configurations for v2beta1
      targetEndpointMetricsV2beta1: |
        - type: Resource
          resource:
            name: cpu
            targetAverageUtilization: 50
    
      # HPA version. For custom metrics HPA version should be v2beta2. Default-> v2beta1
      hpaVersion: "v2beta1"
      ```
    - Depending on your requirements and infrastructure availability, you may change the above values.
- Once you done these changed you have to execute the following command to apply these changed in the Kubernetes cluster.
    ```sh
    >> kubectl apply -f api-operator/controller-configs/controller_conf.yaml
    ```

#### How to configure the default security

- API Operator provides the security to your API. You need to define the security which needs to be applied on the API under "security" extension in the API.

- At the time of the installation of the API Operator the default security will be applied. If you need to change the default security configurations either you can edit the security custom resource or you can edit the file within the distribution.

- If you need to apply the same security configurations to all the APIs then you need to change the default security configurations. If you prefer any other security type as the default security type, you may need to change the values of the default security configuration. 

- The security policy does not need to be mentioned in the swagger and the default security configurations will be applied automatically. 

- Default security configurations are in the ***api-operator/controller-configs/default_security_cr.yaml*** file.

- Default configurations are shown below.
    ```yaml
    apiVersion: wso2.com/v1alpha1
    kind: Security
    metadata:
      name: default-security-jwt
      namespace: wso2-system
    spec:
      type: JWT
      securityConfig:
        - certificate: wso2am320-secret
          issuer: https://wso2apim:32001/oauth2/token
          audience: http://org.wso2.apimgt/gateway
          validateSubscription: false
    
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: wso2am320-secret
      namespace: wso2-system
    data:
      server.pem: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tDQpNSUlEcVRDQ0FwR2dBd0lCQWdJRVhiQUJvekFOQmdrcWhraUc5dzBCQVFzRkFEQmtNUXN3Q1FZRFZRUUdFd0pWDQpVekVMTUFrR0ExVUVDQXdDUTBFeEZqQVVCZ05WQkFjTURVMXZkVzUwWVdsdUlGWnBaWGN4RFRBTEJnTlZCQW9NDQpCRmRUVHpJeERUQUxCZ05WQkFzTUJGZFRUekl4RWpBUUJnTlZCQU1NQ1d4dlkyRnNhRzl6ZERBZUZ3MHhPVEV3DQpNak13TnpNd05ETmFGdzB5TWpBeE1qVXdOek13TkROYU1HUXhDekFKQmdOVkJBWVRBbFZUTVFzd0NRWURWUVFJDQpEQUpEUVRFV01CUUdBMVVFQnd3TlRXOTFiblJoYVc0Z1ZtbGxkekVOTUFzR0ExVUVDZ3dFVjFOUE1qRU5NQXNHDQpBMVVFQ3d3RVYxTlBNakVTTUJBR0ExVUVBd3dKYkc5allXeG9iM04wTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGDQpBQU9DQVE4QU1JSUJDZ0tDQVFFQXhlcW9aWWJRL1NyOERPRlErL3FiRWJDcDZWemI1aHpIN29hM2hmMkZaeFJLDQpGMEg2YjhDT016ejgrMG12RWRZVnZiLzMxak1FTDJDSVFoa1FSb2wxSXJ1RDZuQk9ta2p1WEpTQmZpY2tsTWFKDQpaT1JodUNyQjRyb0h4em9HMTlhV21zY0EwZ25mQktvMm9HWFNqSm1uWnhJaCsyWDZzeUhDZnlNWlowMEx6RHlyDQpnb1hXUVh5RnZDQTJheDU0czdzS2lIT00zUDRBOVc0UVV3bW9FaTRIUW1QZ0pqSU00ZUdWUGgwR3RJQU5OK0JPDQpRMUtrVUk3T3p0ZUhDVEx1M1ZqeE0wc3c4UVJheVpkaG5pUEYrVTluM2ZhMW1PNEtMQnNXNG1ETGpnOFIvSnVBDQpHVFgvU0VFR2owQjVIV1FBUDZteXhLRnoyeHdEYUNHdlQrcmR2a2t0T3dJREFRQUJvMk13WVRBVUJnTlZIUkVFDQpEVEFMZ2dsc2IyTmhiR2h2YzNRd0hRWURWUjBPQkJZRUZFRHBMQjRQRGd6c2R4RDJGVjNyVm5Pci9BMERNQjBHDQpBMVVkSlFRV01CUUdDQ3NHQVFVRkJ3TUJCZ2dyQmdFRkJRY0RBakFMQmdOVkhROEVCQU1DQlBBd0RRWUpLb1pJDQpodmNOQVFFTEJRQURnZ0VCQUU4SC9heEFnWGp0OTNIR0NZR3VtVUxXMmxLa2dxRXZYcnlQMlFrUnBieVFTc1RZDQpjTDdaTFNWQjdNVlZIdElzSGg4ZjFDNFhxNlF1OE5VcnF1NVpMQzFwVUJ5YXFSMlpJemNqL09XTEdZUmpTVEhTDQpWbVZJcTlRcUJxMWo3cjZmM0JXcWFPSWlrbm1UekV1cUlWbE9UWTBnTytTSGRTNjJ2cjJGQ3o0eU9yQkV1bEdBDQp2b21zVThzcWc0UGhGbmtoeEk0TTkxMkx5KzJSZ045TDdBa2h6SytFelhZMS9RdGxJL1Z5c05mUzZ6ckhhc0t6DQo2Q3JLS0NHcVFuQm5TdlNUeUY5T1I1S0ZIbmtBd0U5OTVJWnJjU1FpY014c0xoVE1VSERMUS9nUnl5N1YvWnBEDQpNZkFXUis1T2VRaU5BcC9iRzRmakpvVGRvcWt1bDUxKzJiSEhWclU9DQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tDQo=
    type: Opaque
     ```
    - If you have changed the default security custom resource then you need to re apply it after doing the modifications using the following command
    ```sh
   >> apictl apply -f <k8s-api-operator-home>/api-operator/deploy/controller-configs/default_security_cr.yaml
     ```
    - If you have distinct security policies for different APIs you need to create multiple security policies with the required properties.
    - If you prefer any other security type as the default security type, you may need to change the above values.
    - For more information refer [how to define security guide](../HowToGuide/OverviewOfCrds/apply-security-to-api.md)
    

> Did not find what you were looking for? Please let us know by creating a [GitHub issue](https://github.com/wso2/k8s-api-operator/issues).