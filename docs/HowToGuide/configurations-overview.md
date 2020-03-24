### Configuration Overview

##### How to configure Docker-Hub credentials to push the API docker images

- API Operator keeps the built API gateway images for your API definition in Docker Hub repository so that if you wish to add that API to another environment, it would not recreate the API gateway from the scratch.

- For this, Docker-Hub credentials are required.

- You have to configure these credentials in 2 files.

    1. Open \<api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml

        - Update the ***\<username-docker-registry>*** with the Docker-Hub username.
        ```$xslt
        #docker registry name which the mgw image to be pushed.  eg->  dockerRegistry: username
        dockerRegistry: <username-docker-registry>
        ```
        - Example: If your Docker-Hub username is "sampleuser" :
        ```$xslt
        #docker registry name which the mgw image to be pushed.  eg->  dockerRegistry: username
        dockerRegistry: sampleuser
        ```
        - Once you done these changed you have to execute the following command to apply these changed in the Kubernetes cluster.
        ```$xslt
        kubectl apply -f apim-operator/controller-configs/controller_conf.yaml
        ```
    
    2. Open \<api-k8s-crds-home>apim-operator/controller-configs/docker_secret_template.yaml

        - Enter the base 64 encoded username and password of the Docker-Hub account in the following section.
    
        ```$xslt
        data:
         username: ENTER YOUR BASE64 ENCODED USERNAME
         password: ENTER YOUR BASE64 ENCODED PASSWORD
        ```
        - Example: If the username and password are "admin".
        ```$xslt
        data:
         username: YWRtaW4=
         password: YWRtaW4=
        ```
        - Once you done these changed you have to execute the following command to apply these changed in the Kubernetes cluster.
        ```$xslt
        kubectl apply -f apim-operator/controller-configs/controller_conf.yaml
        ```

##### How to configure Readiness and Liveness probes

- Readiness and Liveness probes are responsible to check the health status of your deployment and pods.

- Readiness probe describe if the pod is ready to accept the traffic.

- Liveness probe describe the health status of the pod.

- Depending on you environment, you might want to change these values.

- Open \<api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml

- Following are the default values present in the configuration file.

- Depending on your environment, you may change the values.
    ```$xslt
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
    ```$xslt
    kubectl apply -f apim-operator/controller-configs/controller_conf.yaml
    ```
##### How to change HPA(Horizontal Pod Autoscaler) related configurations

- API Operator provides the HPA capability to the deployed API.
- HPA will be populated from the default values.
- These configurations reside in the \<api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml
    - Find the default values below.
    ```$xslt
      #Maximum number of replicas for the Horizontal Pod Auto-scale. Default->  hpaMaxReplicas: "5"
      hpaMaxReplicas: "5"
      #Avg CPU utliazation(%) to spin up the next pod.  Default->  hpaTargetAverageUtilizationCPU: "50"
      hpaTargetAverageUtilizationCPU: "50"
      #Required CPU usage for pods.   Default-> resourceRequestCPU: "1000m"
      resourceRequestCPU: "1000m"
      #Required Memory usage pods can use.   Default->  resourceRequestMemory: "512Mi"
      resourceRequestMemory: "512Mi"
      #Max CPU usage limit a pod can use.   Default->  resourceLimitCPU: "2000m"
      resourceLimitCPU: "2000m"
      #Max Memory usage limit a pod can use.   Default->  resourceLimitMemory: "512Mi"
      resourceLimitMemory: "512Mi"
    ```
    - Depending on your requirements and infrastructure availability, you may change the above values.
- Once you done these changed you have to execute the following command to apply these changed in the Kubernetes cluster.
    ```$xslt
    kubectl apply -f apim-operator/controller-configs/controller_conf.yaml

##### How to configure the default security

- API Operator provides the security to your API. You need to define the security which needs to be applied on the API under "security" extension in the API.

- But if you have not specified the security in the API definition, the API operator will apply the default security for the API.

- Default security configurations are in the ***apim-operator/controller-configs/default_security_cr.yaml*** file.

- Default configurations are shown below.
    ```$xslt
   apiVersion: wso2.com/v1alpha1
   kind: Security
   metadata:
     name: default-security-jwt
     namespace: wso2-system
   spec:
     type: JWT
     certificate: wso2am310-secret
     issuer: https://wso2apim:32001/oauth2/token
     audience: http://org.wso2.apimgt/gateway
   ---
   apiVersion: v1
   kind: Secret
   metadata:
     name: wso2am310-secret
     namespace: wso2-system
   data:
     server.pem: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tDQpNSUlEcVRDQ0FwR2dBd0lCQWdJRVhiQUJvekFOQmdrcWhraUc5dzBCQVFzRkFEQmtNUXN3Q1FZRFZRUUdFd0pWDQpVekVMTUFrR0ExVUVDQXdDUTBFeEZqQVVCZ05WQkFjTURVMXZkVzUwWVdsdUlGWnBaWGN4RFRBTEJnTlZCQW9NDQpCRmRUVHpJeERUQUxCZ05WQkFzTUJGZFRUekl4RWpBUUJnTlZCQU1NQ1d4dlkyRnNhRzl6ZERBZUZ3MHhPVEV3DQpNak13TnpNd05ETmFGdzB5TWpBeE1qVXdOek13TkROYU1HUXhDekFKQmdOVkJBWVRBbFZUTVFzd0NRWURWUVFJDQpEQUpEUVRFV01CUUdBMVVFQnd3TlRXOTFiblJoYVc0Z1ZtbGxkekVOTUFzR0ExVUVDZ3dFVjFOUE1qRU5NQXNHDQpBMVVFQ3d3RVYxTlBNakVTTUJBR0ExVUVBd3dKYkc5allXeG9iM04wTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGDQpBQU9DQVE4QU1JSUJDZ0tDQVFFQXhlcW9aWWJRL1NyOERPRlErL3FiRWJDcDZWemI1aHpIN29hM2hmMkZaeFJLDQpGMEg2YjhDT016ejgrMG12RWRZVnZiLzMxak1FTDJDSVFoa1FSb2wxSXJ1RDZuQk9ta2p1WEpTQmZpY2tsTWFKDQpaT1JodUNyQjRyb0h4em9HMTlhV21zY0EwZ25mQktvMm9HWFNqSm1uWnhJaCsyWDZzeUhDZnlNWlowMEx6RHlyDQpnb1hXUVh5RnZDQTJheDU0czdzS2lIT00zUDRBOVc0UVV3bW9FaTRIUW1QZ0pqSU00ZUdWUGgwR3RJQU5OK0JPDQpRMUtrVUk3T3p0ZUhDVEx1M1ZqeE0wc3c4UVJheVpkaG5pUEYrVTluM2ZhMW1PNEtMQnNXNG1ETGpnOFIvSnVBDQpHVFgvU0VFR2owQjVIV1FBUDZteXhLRnoyeHdEYUNHdlQrcmR2a2t0T3dJREFRQUJvMk13WVRBVUJnTlZIUkVFDQpEVEFMZ2dsc2IyTmhiR2h2YzNRd0hRWURWUjBPQkJZRUZFRHBMQjRQRGd6c2R4RDJGVjNyVm5Pci9BMERNQjBHDQpBMVVkSlFRV01CUUdDQ3NHQVFVRkJ3TUJCZ2dyQmdFRkJRY0RBakFMQmdOVkhROEVCQU1DQlBBd0RRWUpLb1pJDQpodmNOQVFFTEJRQURnZ0VCQUU4SC9heEFnWGp0OTNIR0NZR3VtVUxXMmxLa2dxRXZYcnlQMlFrUnBieVFTc1RZDQpjTDdaTFNWQjdNVlZIdElzSGg4ZjFDNFhxNlF1OE5VcnF1NVpMQzFwVUJ5YXFSMlpJemNqL09XTEdZUmpTVEhTDQpWbVZJcTlRcUJxMWo3cjZmM0JXcWFPSWlrbm1UekV1cUlWbE9UWTBnTytTSGRTNjJ2cjJGQ3o0eU9yQkV1bEdBDQp2b21zVThzcWc0UGhGbmtoeEk0TTkxMkx5KzJSZ045TDdBa2h6SytFelhZMS9RdGxJL1Z5c05mUzZ6ckhhc0t6DQo2Q3JLS0NHcVFuQm5TdlNUeUY5T1I1S0ZIbmtBd0U5OTVJWnJjU1FpY014c0xoVE1VSERMUS9nUnl5N1YvWnBEDQpNZkFXUis1T2VRaU5BcC9iRzRmakpvVGRvcWt1bDUxKzJiSEhWclU9DQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tDQo=
   type: Opaque
     ```

    - If you prefer any other security type as the default security type, you may need to change the above values.
    - For more information refer [how to define security guide](../HowToGuide/apply-security-to-api.md)
    

| Did not find what you were looking for? Please let us know by creating a [GitHub issue](https://github.com/wso2/k8s-apim-operator/issues).