# API Operator for Kubernetes

## Introduction

As microservices are increasingly being deployed on Kubernetes, the need to expose these microservices as well
documented, easy to consume, managed APIs is becoming important to develop great applications.
The API operator for Kubernetes makes APIs a first-class citizen in the Kubernetes ecosystem.
Similar to deploying microservices, you can now use this operator to deploy APIs for individual microservices or
compose several microservices into individual APIs. With this users will be able to expose their microservice
as managed API in Kubernetes environment without any additional work.


![Alt text](docs/images/K8s-API-Operator.png?raw=true "K8s API Operator")

## Quick Start Guide with Microgateway

Follow the [Quick Start with Kubernetes](https://apim.docs.wso2.com/en/next/publish/api-microgateway/quick-start-with-kubernetes/)
to work with K8s API Operator.

## Standalone Installation

Use the following command to install API Operator

```sh
>> kubectl apply -f https://github.com/wso2/k8s-api-operator/releases/download/v2.0.0-beta/api-operator-configs.yaml
```
