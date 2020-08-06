## Overview API Operator

API Operator provides a fully automated experience for cloud-native API management.
A user can expose an already deployed microservice as an **API** using the API Operator by providing the API definition
of the particular microservice.


Once the API is deployed, it will be deployed as a **managed API**. 

![Alt text](../images/overview.png?raw=true "Title")


For this, API Operator introduced four new custom resource definitions(CRDs) related to the API management domain.

### 1. Custom Resource: Security

**Security** holds security-related information. You can see the API definition and data structure for Security` here.
Security supports different security types: basic-auth, OAuth2, JWT, etc.
The following YAML shows a sample payload for Security with JWT.

```yaml
apiVersion: wso2.com/v1alpha1
kind: Security
metadata:
  name: petstorejwt
spec:
  # Security - JWT
  type: JWT
  securityConfig:
    - issuer: https://wso2apim:32001/oauth2/token
      audience: http://org.wso2.apimgt/gateway
      #create secret with certificate and add secret name
      certificate: wso2am320-secret
      validateSubscription: false
```

### 2. Custom Resource: RateLimiting

**RateLimiting** holds rate-limiting related information. You can see the API definition and data structure for
**RateLimiting** here. The following YAML shows sample payload.

```yaml
apiVersion: wso2.com/v1alpha1
kind: RateLimiting
metadata:
  name: app1
  namespace: app1-ns
spec:
  type: application             
  description: Allow 4 requests per minute   # optional
  timeUnit: min
  unitTime: 1
  requestCount:
    limit: 4
```

### 3. Custom Resource: TargetEndpoint

**TargetEndpoint** holds endpoint related information. You can see the API definition and data for TargetEndpoint here.
API gateway can be deployed in three patterns: **shared**, **private-jet**, and **sidecar**.

If your backend is already running and you need to expose it via a WSO2 Microgateway, you can define the target URL in
the Swagger itself. If your backend service is not running, but you plan to run it in the same Kubernetes cluster,
you can use **TargetEndpoint** with its relevant Docker image. Then APIM Operator will spin-up the corresponding
Kubernetes deployment for the defined backend service itself with the WSO2 Microgateway. 

In **shared** and **private-jet** mode, the backend can be running in separate PODs, but in **sidecar** mode,
the gateway will run in the same POD adjacent to the backend service.
The following YAML shows a sample payload for Target endpoint.

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

### 4. Custom Resource: API
`API` holds API-related information. You can see the API definition and data structure for API  here.
API takes the Swagger definition as a configMap along with replica count and micro-gateway deployment mode.
The following YAML shows sample payload for API.

```yaml
apiVersion: wso2.com/v1alpha1
kind: API
metadata:
  name: online-store
  namespace: default
spec:
  definition:
    interceptors: {}
    swaggerConfigmapNames:
      - online-store-1-swagger
      - pets-2-swagger
    type: swagger
  mode: privateJet
  replicas: 1
  override: false
  version: v1.0.0
  environmentVariables:
    - ENV=DEV
  ingressHostname: mgw.ingress.wso2.com
```

Each of the above CRDs has corresponding custom controllers. Custom controllers are the “brains” behind the custom resources. 

### Custom Controller: Security

![Alt text](../images/security-crd.png?raw=true "Title")

The security controller will store user-defined security policies corresponding to the Security API and creates a
Security secret. It supports JWT, Oauth2, and basic security types out-of-the-box.
When running the Kaniko job by the API controller, it will add to the keystore and then the keystore will be added
to the WSO2 Microgateway Docker image. 

### Custom Controller: RateLimiting

![Alt text](../images/ratelimiting-crd.png?raw=true "Title")


The RateLimiting controller will store the user-defined policy corresponding to the RateLimit API in addition to
default policies provided out-of -the box. It also creates policy template configMaps.
When a new rate limiting policy is added, we update that policy template config map.
When running the Kaniko job by the API controller, it takes this policy template configmap and uses it to build
the Docker image. 

### Custom Controller: TargetEndpoint

![Alt text](../images/targetendpoint-crd.png?raw=true "Title")

The TargetEndpoint controller will store target endpoint metadata corresponding to the TargetEndpoint API.
If the mode of the target endpoint is **privateJet**, it will create Deployment, Service and PODs for
relevant backend services. If the mode is **sidecar**, it will store the definition and when we add a micro gateway
with this endpoint, it will create PODs with the gateway attached as a sidecar to the service. 

### Custom Controller: API

![Alt text](../images/api-crd.png?raw=true "Title")


API controller is quite complex compared to other controllers. It has two main tasks.  
- Build an API Microgateway container and push it to the registry configured during the API operator installation.
- Create Kubernetes artifacts and deploy them into Kubernetes clusters.

When the API custom controller is triggered, it will receive a Swagger definition from the attached configMap and
creates a Kaniko job by attaching a multi-step Dockerfile along with the Swagger definition.
This Dockerfile is using pre-build the Docker image that has the WSO2 Microgateway toolkit.
WSO2 Microgateway toolkit will generate the API Microgateway runtime with the corresponding swagger file passed.
Finally Kaniko build create a new API Microgateway docker image and push to the configured docker registry.

After finishing the step one, API controller will start creating relevant Kubernetes artifacts corresponding to the 
API definition. Depending on defined API mode, it will create Kubernetes deployment for both API microgateway and
backend services. 

As you can see, API Controller has taken out all the complexity from DevOps and automates deployment with all
the best practices required to deploy API Microgateway along with microservices architecture.
