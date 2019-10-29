### Configuration Overview

##### How to configure Docker-Hub credentials to push the API docker images

- API Operator keeps the built API gateway images for your API definition in Docker Hub repository so that if you wish to add that API to another environment, it would not recreate the API gateway from the scratch.

- For this, Docker-Hub credentials are required.

- You have to configure these credentials in 2 files.

    1. Open <api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml

        - Update the ***<username-docker-registry>*** with the Docker-Hub username.
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
    
    2. Open <api-k8s-crds-home>apim-operator/controller-configs/docker_secret_template.yaml

        - Enter the base 64 encoded username and password of the Docker-Hub account in the following section.
    
        ```$xslt
        data:
         username: ENTER YOUR BASE64 ENCODED USERNAME
         password: ENTER YOUR BASE64 ENCODED PASSWORD
        ```
        - Example: IF the username and password are "admin".
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

- Open <api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml

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
- These configurations reside in the <api-k8s-crds-home>/apim-operator/controller-configs/controller_conf.yaml
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

- Default security configurations are in the ***api-k8s-crds-1.0.0/apim-operator/controller-configs/default_security_cr.yaml*** file.

- Default configurations are shown below.
    ```$xslt
   apiVersion: wso2.com/v1alpha1
   kind: Security
   metadata:
     name: default-security-jwt
     namespace: wso2-system
   spec:
     type: JWT
     certificate: wso2am300-secret
     issuer: https://wso2apim:32001/oauth2/token
     audience: http://org.wso2.apimgt/gateway
   ---
   apiVersion: v1
   kind: Secret
   metadata:
     name: wso2am300-secret
     namespace: wso2-system
   data:
     server.pem: MIIDfTCCAmWgAwIBAgIEbfVjBzANBgkqhkiG9w0BAQsFADBkMQswCQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDU1vdW50YWluIFZpZXcxDTALBgNVBAoTBFdTTzIxDTALBgNVBAsTBFdTTzIxEjAQBgNVBAMTCWxvY2FsaG9zdDAeFw0xOTA4MjMxMjUwMzNaFw0yOTA4MjAxMjUwMzNaMGQxCzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNTW91bnRhaW4gVmlldzENMAsGA1UEChMEV1NPMjENMAsGA1UECxMEV1NPMjESMBAGA1UEAxMJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAj3mYMT8N2SR8cpimdMOTpk/M8fOxPF1BHQAiCtld4nbksILgJKsA34GSP5Oh4gLW21VCEPPzdGLnqfwM6ZoG/X0rcK5++VbqH/vH4Cclba6fqlLxCvTiRbPJ58Pe7+biCeQ368dG2aeBPV3EhO8br3Z/LcQXASmhSWps8J3GOSx/49xzkHh59J2gJHhnvjxcszZAF35SLAb6F+2rJQrOJs6u7WfJv4NQxSyhcgcr4/+77JzNFEVUj4TPSBy2WGAgK5ttP5+kG3+rKs0lQjTo9h/hK89KjbbRvoqZdpxnwQYxFDOk0CxijZVO5Cs3cabeUHZeXehHSgXj6W+VGMiDgwIDAQABozcwNTAUBgNVHREEDTALgglsb2NhbGhvc3QwHQYDVR0OBBYEFFU2A4pBuR0aKGrgQAtrSlqWrNLLMA0GCSqGSIb3DQEBCwUAA4IBAQAX+F30hIwI+8hO9IQ9Wr40+zL6KTgDxWraB450D7UyZ/FApKK2R/vYvIqab+H6u9XNCz63+sYgX6/UBSYW47ow6QMcv71xepXbwtLqq3MQr6frgP52Z2jyQtAbDpirh4/IXkhF+S8DsDFxmlPy423LKnTqCqIfyv7Y8Y8lty5BWyfYJV7V2RJnZ4zIKv66U3exxugR0WRGWy56nIY8GGaroxuC9cH6NkVwN9GmYoCa9PUGynQ4NHjeg6VSwQZ279VGpLhogWS67x8V/nR+yjI+qTjjCbJqsoHVQL90Vxa+ASD1DViBA8ar1/9Ns5vIEZet5GT1nM10ZzEK+E1QMGed 
    ```

    - If you prefer any other security type as the default security type, you may need to change the above values.
    - For more information refer [how to define security guide](../HowToGuide/apply-security-to-api.md)
    

| Did not find what you were looking for? Please let us know by creating a [GitHub issue](https://github.com/wso2/k8s-apim-operator/issues).