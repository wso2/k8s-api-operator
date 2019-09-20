# WSO2 API Manager Operator for Kubernetes

## Introduction

WSO2 API Manager is a full lifecycle API Management solution with monetization and policy enforcement. WSO2 API Microgateway is a cloud native, developer centric and decentralized API gateway for microservices. Kubernetes (K8s) is an open-source system for automating deployment, scaling, and management of containerized applications. The intention of this project is to provide cloud native full API management by a seamless integration with Kubernetes. With this native API Management support on K8s, this targets to improve the developer/devOps experience.

## Approach

#### Developer First Approach In WSO2 API Microgateway

![alt text](https://raw.githubusercontent.com/wso2/k8s-apim-operator/master/api-microgateway.png)

Open API definition is considered as the single source of truth to the WSO2 API Microgateway. This Open API definition contains all the required information regarding your API. By providing this definition to the WSO2 API Microgateway, you can generate a balx file which is required to deploy your API in WSO2 API Microgateway. 

#### API Manager Operator for Kubernetes

![alt text](https://raw.githubusercontent.com/wso2/k8s-apim-operator/master/apim-operator.png)

The developer first approach is used when creating the API Manager Operator for Kubernetes. When an user requires to expose an API for the service he created, he only needs to provide the Open API definition to the Kubernetes. Then it will create the API and deploy his API in the WSO2 API Microgateway. His API is exposed as the Load Balancer service type in Kubernetes. 

#### API Manager Custom Resources for Kubernetes

We have initially introduced four custom resources for Kubernetes.

- API <br>
  Holds API related information
  
- Target Endpoint <br>
  Holds endpoint related information
    
- Security <br>
  Holds security related information

- Rate Limiting <br>
  Holds rate limiting related information

#### Kubernetes CLI(kubectl) plugins 


We have introduced this feature with [APIM CLI](https://github.com/wso2/product-apim-tooling/releases) tool to deploy and manger APIs and related services in kubernetes cluster without any hassle.

Alternative:

We have two kubectl plugins which helps to add an API and update an API. As part of installing the kubectl, users will have to install these plugins.

---

## Quick Start Guide

##### Step 1: Install [Kubernetes v1.12 or above](https://kubernetes.io/docs/setup/)

##### Step 2: Download [wso2am-k8s-crds-v0.8-alpha.zip](https://github.com/wso2/k8s-apim-operator/releases) and extract the zip

1. This zip contains the artifacts that required to deploy in Kubernetes.
2. Extract wso2am-k8s-crds-1.0.zip and navigate to the \<APIM-K8s-CRD-HOME>/ directory.
```
cd <APIM-K8s-CRD-HOME>/
```
   
**Note:** You need to run all commands from within the <APIM-K8s-CRD-HOME>/ directory.

##### Step 3: Configure APIM CLI tool
- Navigate to the API Management Tooling page - https://github.com/wso2/product-apim-tooling/releases
- Download tooling archive suitable for your platform (i.e., Mac, Windows, Linux) and extract it the CLI tool that you downloaded to a desired location and cd into it.

- Navigate to the working directory where the executable CLI Tool resides.

- Execute the following command to start the CLI tool.

```
./apimcli
```

Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

For further instructions execute the following command.
```
apimcli --help
```
Set the APIM CLI tool's mode to kubernetes or k8s to be compatible with kubectl commands

```$xslt
apimcli set --mode k8s
```
 or 
```
apimcli set --mode kubernetes
```

###### Alternative (without cli)
- Give executable permission to the extension files

**Note: It is highly recommend to use the apimcli approach instead of going ahead with kubectl extensions**
```
chmod +x ./deploy/kubectl-extension/kubectl-add
chmod +x ./deploy/kubectl-extension/kubectl-update
```

- Copy the extensions to ***/usr/local/bin/***
```
cp ./deploy/kubectl-extension/kubectl-add /usr/local/bin
cp ./deploy/kubectl-extension/kubectl-update /usr/local/bin
```


##### Step 4: Deploy K8s CRD artifacts

- Deploying CRDs for API, TargetEndpoint, Security, RateLimiting
```
apimcli apply -f ./deploy/crds/
```

- Deploying namespace, roles/role binding and service account associated with the operator
```
apimcli apply -f ./deploy/controller-artifacts/
```

- Deploying controller level configurations

"controller-configs" contains the configuration user needs to change. The docker images are created and pushed to the user's docker registry.
Update the ***user's docker registry*** in the controller_conf.yaml. Enter the base 64 encoded username and password of the user's docker registry into the docker_secret_template.yaml.

```
apimcli apply -f ./deploy/controller-configs/
```

##### Step 5: Deploy an API in K8s cluster

- Deploy the API
```
apimcli add api -n "api_name" --from-file="location to the api swagger definition"

apimcli add api -n petstore --from-file=./deploy/scenarios/scenario-1/petstore_basic.yaml
```
  
- Update the API
```
apimcli update api -n "api_name" --from-file="location to the api swagger definition"

apimcli update api -n petstore --from-file=./deploy/scenarios/scenario-1/petstore_basic.yaml
```
  
- Delete the API
```
apimcli delete api "api_name"

apimcli delete api petstore
```

Optional Parameters

```
--replicas=3          Number of replicas
--namespace=wso2      Namespace to deploy the API

apimcli add api -n "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" --namespace="desired namespace"
```

**Note:** Namespace and replicas are optional parameters. If they are not provided default namespace will be used and 1 replica will be created. However, the namespace used in all the commands related to particular API name must match.

##### Step 6: Invoke the Petstore API

###### Step 6.1: Obtain a token

After the APIs are exposed via WSO2 API Microgateway, you can invoke an API with a valid JWT token or an opaque access token. In order to use JWT tokens, WSO2 API Microgateway should be presented with a JWT signed by a trusted OAuth2 service.
Let's use the following sample JWT token for the quick start guide. Here we will be using an never expiring jwt token acquired from WSO2 API Manager.

Sample Token
```
eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UQXhabU14TkRNeVpEZzNNVFUxWkdNME16RXpPREpoWldJNE5ETmxaRFUxT0dGa05qRmlNUSJ9.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbiIsImFwcGxpY2F0aW9uIjp7ImlkIjoyLCJuYW1lIjoiSldUX0FQUCIsInRpZXIiOiJVbmxpbWl0ZWQiLCJvd25lciI6ImFkbWluIn0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC9sb2NhbGhvc3Q6OTQ0M1wvb2F1dGgyXC90b2tlbiIsImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6Ilg5TGJ1bm9oODNLcDhLUFAxbFNfcXF5QnRjY2EiLCJleHAiOjM3MDMzOTIzNTMsImlhdCI6MTU1NTkwODcwNjk2MSwianRpIjoiMjI0MTMxYzQtM2Q2MS00MjZkLTgyNzktOWYyYzg5MWI4MmEzIn0=.b_0E0ohoWpmX5C-M1fSYTkT9X4FN--_n7-bEdhC3YoEEk6v8So6gVsTe3gxC0VjdkwVyNPSFX6FFvJavsUvzTkq528mserS3ch-TFLYiquuzeaKAPrnsFMh0Hop6CFMOOiYGInWKSKPgI-VOBtKb1pJLEa3HvIxT-69X9CyAkwajJVssmo0rvn95IJLoiNiqzH8r7PRRgV_iu305WAT3cymtejVWH9dhaXqENwu879EVNFF9udMRlG4l57qa2AaeyrEguAyVtibAsO0Hd-DFy5MW14S6XSkZsis8aHHYBlcBhpy2RqcP51xRog12zOb-WcROy6uvhuCsv-hje_41WQ==
```

###### Step 6.2: Invoke the API

Execute the command below to set a self-contained OAuth2.0 access token in the JWT format as a variable on your terminal session.

```
TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UQXhabU14TkRNeVpEZzNNVFUxWkdNME16RXpPREpoWldJNE5ETmxaRFUxT0dGa05qRmlNUSJ9.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbiIsImFwcGxpY2F0aW9uIjp7ImlkIjoyLCJuYW1lIjoiSldUX0FQUCIsInRpZXIiOiJVbmxpbWl0ZWQiLCJvd25lciI6ImFkbWluIn0sInNjb3BlIjoiYW1fYXBwbGljYXRpb25fc2NvcGUgZGVmYXVsdCIsImlzcyI6Imh0dHBzOlwvXC9sb2NhbGhvc3Q6OTQ0M1wvb2F1dGgyXC90b2tlbiIsImtleXR5cGUiOiJQUk9EVUNUSU9OIiwic3Vic2NyaWJlZEFQSXMiOltdLCJjb25zdW1lcktleSI6Ilg5TGJ1bm9oODNLcDhLUFAxbFNfcXF5QnRjY2EiLCJleHAiOjM3MDMzOTIzNTMsImlhdCI6MTU1NTkwODcwNjk2MSwianRpIjoiMjI0MTMxYzQtM2Q2MS00MjZkLTgyNzktOWYyYzg5MWI4MmEzIn0=.b_0E0ohoWpmX5C-M1fSYTkT9X4FN--_n7-bEdhC3YoEEk6v8So6gVsTe3gxC0VjdkwVyNPSFX6FFvJavsUvzTkq528mserS3ch-TFLYiquuzeaKAPrnsFMh0Hop6CFMOOiYGInWKSKPgI-VOBtKb1pJLEa3HvIxT-69X9CyAkwajJVssmo0rvn95IJLoiNiqzH8r7PRRgV_iu305WAT3cymtejVWH9dhaXqENwu879EVNFF9udMRlG4l57qa2AaeyrEguAyVtibAsO0Hd-DFy5MW14S6XSkZsis8aHHYBlcBhpy2RqcP51xRog12zOb-WcROy6uvhuCsv-hje_41WQ==
```

The API service is exposed as the Load Balancer service type. You can get the service endpoint details by using the following command.

```
apimcli get services
```

Sample Output:

```
NAME          TYPE         CLUSTER-IP      EXTERNAL-IP          PORT(S)                     AGE
petstore   LoadBalancer    10.0.3.74     104.199.77.249   9095:30453/TCP,9090:32422/TCP     1m
```

You can now invoke the API running on the Microgateway using cURL as below

Format
```
curl -X GET "<EXTERNAL-IP>:<MICROGATEWAY-PORT>/<API-context>/<API-resource>" -H "accept: application/xml" -H "Authorization:Bearer <JWT_TOKEN>" -k
```

Examples

```
curl -X GET "https://104.199.77.249:9095/petstore/v1/pet/findByStatus?status=available" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
 
curl -X GET "https://104.199.77.249:9095/petstore/v1/pet/1" -H "accept: application/xml" -H "Authorization:Bearer $TOKEN" -k
```

##### Cleanup

```
apimcli delete -f ./deploy/controller-configs/
apimcli delete -f ./deploy/controller-artifacts/
apimcli delete -f ./deploy/crds/
```
##### Deploying APIM and APIM Analytics in K8s Cluster

Kubernetes artifacts to deploy APIM and APIM analytics deployment are shipped with the this distribution.

Navigate to wso2am-k8s-crds-v0.8-alpha/apim-operator/apim-deployment

- Deploy API Manager in Kubernetes Cluster

```$xslt
apimcli apply -f api-manager
```
- Deploy APIM Analytics in Kubernetes Cluster
```$xslt
apimcli apply -f analytics
```
    
##### Sample Scenarios

1. [Sample 1: Basic Petstore Sample](scenarios/scenario-1)
1. [Sample 2: Secure an API with basic authentication](scenarios/scenario-2)
1. [Sample 3: Secure an API with JWT](scenarios/scenario-3)
1. [Sample 4: Secure an API with OAuth2 tokens](scenarios/scenario-4)
1. [Sample 5: Apply rate limiting for an API](scenarios/scenario-5)
1. [Sample 6: Private jet mode for API and Endpoint](scenarios/scenario-6)
1. [Sample 7: Sidecar mode for API and Endpoint](scenarios/scenario-7)

##### Troubleshooting Guide

You can refer [troubleshooting guide](docs/Troubleshooting/troubleshooting.md).