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

#### APIM CLI for Kubernetes CRDs


We have introduced this feature with [APIM CLI](https://github.com/wso2/product-apim-tooling/releases) tool to deploy and manger APIs and related services in kubernetes cluster without any hassle.

##### Deploying APIM in  K8S Cluster

Kubernetes artifacts to deploy APIM and APIM analytics deployment are shipped with the this distribution.

Navigate to wso2am-k8s-crds/apim-operator/apim-deployment/api-manager

- Deploy API Manager in Kubernetes Cluster

```$xslt
apimcli apply -f api-manager/k8s-artifacts
```
- Deploy APIM Analytics in Kubernetes Cluster
```$xslt
apimcli apply -f api-manager/analytics
```

---

## Quick Start Guide

##### Step 1: Install [Kubernetes v1.12 or above](https://kubernetes.io/docs/setup/)

##### Step 2: Download [wso2am-k8s-crds-v1.0.0-beta.zip](https://github.com/wso2/k8s-apim-operator/releases/download/v1.0.0-beta/wso2am-k8s-crds-v1.0.0-beta.zip) and extract the zip

1. This zip contains the artifacts that required to deploy in Kubernetes.
2. Extract wso2am-k8s-crds-1.0-beta.zip and navigate to the \<APIM-K8s-CRD-HOME>/apim-operator directory.
```
cd <APIM-K8s-CRD-HOME>/
```
   
**Note:** You need to run all commands from within the <APIM-K8s-CRD-HOME>/ directory.

##### Step 3: Configure APIM CLI tool
- Navigate to the API Management Tooling page - https://github.com/wso2/product-apim-tooling/releases/
- Download tooling archive (from v3.0.0-beta onwards) suitable for your platform (i.e., Mac, Windows, Linux) and extract it the CLI tool that you downloaded to a desired location and cd into it.

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

##### Step 5: Deploy an API in K8s cluster via CRDs

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

###### Publishing API in the API Manager

Since APIM deployment is already deployed in the k8s cluster (refer  **Deploying APIM in K8S Cluster** in ReadMe), App developers/subscribers can navigate to the devportal (https://wso2apim:9443/devportal) and obtain a JWT access token by subscribing the APIs.
To subscribe the APIs to the application, the API is needed to be published in the API Manager in k8s.

<br>Following commands will help you to publish the API in the API manager.
Using the APIM CLI tool, init the project using the sample swagger file and import that to the API Manager in Kubernetes deployment.
Commands of the CLI can be found [here](https://github.com/wso2/product-apim-tooling/blob/v3.0.0-beta/import-export-cli/docs/apimcli.md)  

Using the APIM CLI command, adding the environment to the CLI configs/
```
apimcli add-env -e k8s --registration https://wso2apim:9443/client-registration/v0.15/register --apim https://wso2apim:9443 --token https://wso2apim:8243/token --admin https://wso2apim:9443/api/am/admin/v0.15 --api_list https://wso2apim:9443/api/am/publisher/v0.15/apis --app_list https://wso2apim:9443/api/am/store/v0.15/applications

```
Init the API project using CLI command

```
apimcli init petstore --oas=./deploy/scenarios/scenario-1/petstore_basic.yaml
```

Import the API to the k8s environment.
(You need to change the API life cycle status before importing, to published in the api.yaml file to publish the API)
```
./apimcli import-api -f petstore/ -e k8s -k 

```


###### Step 6.1: Obtain a token

After the APIs are exposed via WSO2 API Microgateway, you can invoke an API with a valid JWT token or an opaque access token. In order to use JWT tokens, WSO2 API Microgateway should be presented with a JWT signed by a trusted OAuth2 service.
Let's use the following sample JWT token for the quick start guide. Here we will be using an never expiring jwt token acquired from WSO2 API Manager.

Sample Token
```
eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjV9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06OTQ0M1wvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoiOFpWV1lQYkk2Rm1lY0ZoeXdVaDVVSXJaNEFvYSIsImV4cCI6MzcxODI5OTU1MiwiaWF0IjoxNTcwODE1OTA1LCJqdGkiOiJkMGI2NTgwNC05NDk3LTQ5ZjktOTcxNC01OTJmODFiNzJhYjMifQ==.HYCPxCbNcALcd0svu47EqFoxnnBAkVJSnCPnW6jJ1lZQTzSAiuiPcGzTnyP1JHodQknhYsSrvdZDIzWzU_mRH2i3-lMVdm0t43r-0Ti0EdBSX2756ilo266MVeWhxbz9p3hPm5ndDCoo_bfB4KbjigjmhXv_PJyUMuWtMo669sHQNs5FkiOT2X0gzFP1iJUFf-H9y762TEIYpylKedVDzQP8x4LCRZsO54e1iA-DZ5h5MKQhJsbKZZ_MMXGmtdo8refPyTCc7HIuevUXIWAaSNRFYj_HZTSRYhFEUtDWn_tJiySn2umRuP3XqxPmQal0SxD7JiV8DQxxyylsGw9k6g==
```

###### Step 6.2: Invoke the API

Execute the command below to set a self-contained OAuth2.0 access token in the JWT format as a variable on your terminal session.

```
TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IlpqUm1ZVE13TlRKak9XVTVNbUl6TWpnek5ESTNZMkl5TW1JeVkyRXpNamRoWmpWaU1qYzBaZz09In0=.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbkBjYXJib24uc3VwZXIiLCJhcHBsaWNhdGlvbiI6eyJvd25lciI6ImFkbWluIiwidGllciI6IlVubGltaXRlZCIsIm5hbWUiOiJzYW1wbGUtY3JkLWFwcGxpY2F0aW9uIiwiaWQiOjV9LCJzY29wZSI6ImFtX2FwcGxpY2F0aW9uX3Njb3BlIGRlZmF1bHQiLCJpc3MiOiJodHRwczpcL1wvd3NvMmFwaW06OTQ0M1wvb2F1dGgyXC90b2tlbiIsInRpZXJJbmZvIjp7fSwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W10sImNvbnN1bWVyS2V5IjoiOFpWV1lQYkk2Rm1lY0ZoeXdVaDVVSXJaNEFvYSIsImV4cCI6MzcxODI5OTU1MiwiaWF0IjoxNTcwODE1OTA1LCJqdGkiOiJkMGI2NTgwNC05NDk3LTQ5ZjktOTcxNC01OTJmODFiNzJhYjMifQ==.HYCPxCbNcALcd0svu47EqFoxnnBAkVJSnCPnW6jJ1lZQTzSAiuiPcGzTnyP1JHodQknhYsSrvdZDIzWzU_mRH2i3-lMVdm0t43r-0Ti0EdBSX2756ilo266MVeWhxbz9p3hPm5ndDCoo_bfB4KbjigjmhXv_PJyUMuWtMo669sHQNs5FkiOT2X0gzFP1iJUFf-H9y762TEIYpylKedVDzQP8x4LCRZsO54e1iA-DZ5h5MKQhJsbKZZ_MMXGmtdo8refPyTCc7HIuevUXIWAaSNRFYj_HZTSRYhFEUtDWn_tJiySn2umRuP3XqxPmQal0SxD7JiV8DQxxyylsGw9k6g==
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
    
##### Sample Scenarios

1. [Sample 1: Expose a K8s service as an API](scenarios/scenario-1)
1. [Sample 2: Basic Petstore Sample](scenarios/scenario-2)
1. [Sample 3: Secure an API with basic authentication](scenarios/scenario-3)
1. [Sample 4: Secure an API with JWT](scenarios/scenario-4)
1. [Sample 5: Secure an API with OAuth2 tokens](scenarios/scenario-5)
1. [Sample 6: Apply rate limiting for an API](scenarios/scenario-6)
1. [Sample 7: Private jet mode for API and Endpoint](scenarios/scenario-7)
1. [Sample 8: Sidecar mode for API and Endpoint](scenarios/scenario-8)
1. [Sample 9: Expose an API with multiple service endpoints](scenarios/scenario-9)

##### Troubleshooting Guide

You can refer [troubleshooting guide](docs/Troubleshooting/troubleshooting.md).