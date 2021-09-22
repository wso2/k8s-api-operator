# API Operator for Kubernetes

## Introduction

As microservices are increasingly being deployed on Kubernetes, the need to expose these microservices as well
documented, easy to consume, managed APIs is becoming important to develop great applications.
The API operator for Kubernetes makes APIs a first-class citizen in the Kubernetes ecosystem.
Similar to deploying microservices, you can now use this operator to deploy APIs for individual microservices or
compose several microservices into individual APIs. With this users will be able to expose their microservice
as managed API in Kubernetes environment without any additional work. 

The API operator for Kubernetes provides first-class support for Micro Integrator deployments in the Kubernetes
ecosystem. It uses the Integration custom resource (integration_cr.yaml file) that is available in the Kubernetes
project (exported from WSO2 Integration Studio) and deploys the integration in your Kubernetes environment.

![Alt text](images/K8s-API-Operator.png?raw=true "K8s API Operator")

## Quick Start Guide with Choreo Connect

Deploy your first API to Choreo Connect using K8s API Operator, see [Quick Start Guide - Kubernetes](https://apim.docs.wso2.com/en/4.0.0/deploy-and-publish/deploy-on-gateway/choreo-connect/getting-started/quick-start-guide/quick-start-guide-kubernetes/)


## Deploying Integrations

Deploy integrations using K8s API Operator, see [Deploying Integrations using the Operator](https://apim.docs.wso2.com/en/4.0.0/install-and-setup/setup/kubernetes-operators/k8s-api-operator/manage-integrations/integration-deployments/)

## Standalone Installation

Use the following command to install API Operator

```sh
>> kubectl apply -f https://github.com/wso2/k8s-api-operator/releases/download/v2.0.1/api-operator-configs.yaml
```

## Documentation 

You can find the documentation [here](https://apim.docs.wso2.com/en/4.0.0/reference/k8s-operators/k8s-api-operator/).