# API Operator Production Deployment Guide

This documentation will guide you on how to use the Operator on a production scenario.
Here, we will see how we can deploy a simple API which calls the petstore service 
([Scenario-2](../../scenarios/scenario-2)) on two different environments (Development and QA).

### Install API Operator

First, install the API Operator.

- Download [k8s-api-operator-1.2.2.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.2/k8s-api-operator-1.2.2.zip)
and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    
    2. Extract k8s-api-operator-1.2.2.zip
    
    ```sh
    >> cd k8s-api-operator-1.2.2
    ```
 
    **_Note:_** You need to run all commands from within the ***k8s-api-operator-1.2.2*** directory.
 
- Download API controller v3.2.0 or the latest v3.2.x from the [API Manager Tooling web site](https://wso2.com/api-management/tooling/)

    - Under Dev-Ops Tooling section, you can download the tool based on your operating system.

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

- You can find available operations using the below command.
    
  ```sh
  >> apictl --help

- Execute the following command to install API Operator and the configurations interactively on `wso2-system` namespace 
  and configure a repository to push the built
managed API image.
- Select the desired repository type.
- Enter repository name of repository you chose.
- Enter the username and the password.
- Confirm the configurations are correct with entering "Y".

```sh
>>> apictl install api-operator

Choose registry type:
1: Docker Hub
2: Amazon ECR
3: GCR
4: HTTP Private Registry
5: HTTPS Private Registry
6: Quay.io
Choose a number: 1: 1
Enter repository name: johndoe
Enter username: johndoe
Enter password: 

Repository: johndoe
Username  : johndoe
Confirm configurations: Y: y
```

Please refer [Install API Operator in CI/CD](install-api-operator-in-cicd.md) for more information
on installing the API operator and [Working with Registries](WorkingWithDockerRegistries) for more information
on supported registry types.

```sh
namespace/wso2-system created
customresourcedefinition.apiextensions.k8s.io/apis.wso2.com created
customresourcedefinition.apiextensions.k8s.io/ratelimitings.wso2.com created
customresourcedefinition.apiextensions.k8s.io/securities.wso2.com created
customresourcedefinition.apiextensions.k8s.io/targetendpoints.wso2.com created
...
...
[Setting to K8s Mode]
```

### Deploy an API in Development Environment

For the sake of this example, we assume that our development environment is `wso2-dev` namespace.

Now we will deploy the Petstore API. Go to [Scenario-2](../../scenarios/scenario-2) and use the following command
to deploy the API in the Dev namespace.

```sh
>>> apictl add api -n petstore-api --from-file=swagger.yaml \
    --namespace=wso2-dev
```
```sh
creating configmap with swagger definition
configmap/petstore-api-1-swagger created
creating API definition
api.wso2.com/petstore-api created
```

You can use the `apict update api` command to make tweaks to the deployed API by changing the configurations 
in the swagger definition. Operator will create a new image in your docker registry with the update timestamp. 
Alternatively, you can delete the current API and deploy a new API with the `--override` flag. This will override 
the existing image in your docker registry.

### Deploy an API in QA Environment

When deploying the API you tried out in the Dev environment in your QA environment, you can follow any of the 
following options. Here, we assume that our QA environment is `wso2-qa` namespace.

#### 1. Deploying API using Swagger definition

You can use the same swagger definition used in the Dev environment to deploy the APIs in higher environments. 
You can update the swagger definition to suit your QA/Prod needs by changing endpoints, security, etc. 
Operator will create a docker image for each of the environments while having the same source.

```sh
>>> apictl add api -n petstore-api-qa --from-file=swagger.yaml \
    --namespace=wso2-qa
```

#### 2. Deploying API using existing configurations

Following are a few out of many cases you would come across when moving from a lower environment to a higher
environment.

* ##### Case 1 - Single K8s cluster, Dev and QA namespaces, Single registry

    Once the API is created in the Dev environment, the docker image will be pushed to your docker registry by 
    the Operator. We will consider this as the Dev image. Here, we are using an HTTP private registry to store our docker 
    images. Refer [Working with Docker Registries](WorkingWithDockerRegistries) to find out more about the supported 
    registry types.
    
    ```sh
    >>> docker tag 192.163.3.8:5000/operator/petstore-api:v1 192.163.3.8:5000/operator-test/petstore-api-qa:v1
  
    >>> docker push 192.163.3.8:5000/operator/petstore-api-qa:v1
    ```
  
    The above commands will push the QA docker image of the API to the specified registry.
    
    Now you can deploy the API in your QA environment.
    
    ```sh
    >>> apictl add api -n petstore-api-qa --from-file=swagger.yaml --namespace=wso2-qa \ 
        --image=192.163.3.8:5000/operator/petstore-api-qa:v1
    ```
    `--image` flag will use the QA image we pushed to the registry rather than creating the API from the scratch.
    
    ```sh
    creating configmap with swagger definition
    configmap/petstore-api-qa-1-swagger created
    creating API definition
    api.wso2.com/petstore-api-qa created
    ```
  
* ##### Case 2 - Single K8s cluster, Dev and QA namespaces, Different registry
    
    In this case, we are storing the docker images of the APIs of higher environments in a different registry.
    
    ```sh
    >>> docker tag 192.163.3.8:5000/operator-test/petstore-api:v1 wso2qa/petstore-api-qa:v1
    
    >>> docker push wso2qa/petstore-api-qa:v1
    ```
  
    This will push our QA image to wso2qa repository in Dockerhub.
    
    ```sh
    The push refers to repository [docker.io/wso2qa/petstore-api-qa]
    ....
    ....
    ```
    
    Similar to the Case 1, we can deploy the API in our QA environment using the `--image` flag.
    
    ```sh
    >>> apictl add api -n petstore-api-qa --from-file=swagger.yaml --namespace=wso2-qa \
        --image=wso2qa/petstore-api-qa:v1
    ```
  
* ##### Case 3 - Different K8s clusters

    This case is different from previous cases as you have to install the API Operator in the cluster again. You can
    follow the [Install API Operator in CI/CD](install-api-operator-in-cicd.md) guide to deploy the Operator in your
    QA cluster.
    
    Once the Operator is up and running, we can deploy the API using the `--image` flag. This will use the image you
    specify instead of deploying the API from the scratch.
    
    ```sh
    >>> apictl add api -n petstore-api-qa --from-file=swagger.yaml \ 
        --image=wso2qa/petstore-api-qa:v1
    ```
  
    This will deploy the API in the `default` namespace in your QA cluster.
  
#### Changing Endpoints

When moving from Dev to QA environments, you may need to change the backend service endpoint or the targetendpoint 
to something different from the one in Dev environment. There are two ways in which we can handle this situation.

* ##### Approach 1 - Develop a new API with updated swagger definition
    
    In the swagger definition used in this example, we have defined the production ednpoint as follows.
    
    ```yaml
    x-wso2-production-endpoints:
      urls:
        - https://petstore.swagger.io/v2
    ```
  
    You need to change the url to your preferred one before deploying the API in the QA environment. This will push
    a new image of the API to your docker registry with the updated endpoint. You can't follow the above 2 cases if
    you are using this approach.
    
    ```sh
    >>> apictl add api -n petstore-api-qa --from-file=swagger.yaml \ 
        --namespace=wso2-qa
    ```
    
* ##### Approach 2 - Using the environment variables to pass endpoint values

    To pass the endpoint values through environment variables, you need to define the endpoints in a different format
    in the swagger definition of the API you deploy in the Dev environment.
    
    ```yaml
    x-wso2-production-endpoints: "#/x-wso2-endpoints/myEndpoint1"
    x-wso2-endpoints:
      - myEndpoint1:
          urls:
          - https://petstore.swagger.io/v2

    ``` 
    
    Here, the reference made in `x-wso2-production-endpoints` points to the `myEndpoint1` defined under `x-wso2-endpoints`.
    You can follow either Case 1 or Case 2 in this situation.
    
    When you are deploying the API in the QA environment, define the preferred endpoint as an environment variable in 
    the `apictl add api` command.
    
    ```sh
    >>> apictl add api -n petstore-api-qa --from-file=swagger.yaml \
        --image=192.163.3.3:5000/operator-test/petstore-api-qa:v1 \
        --namespace=wso2-qa \
        --env=myEndpoint1_prod_endpoint_0="https://petstore3.swagger.io/api/v3"
    ```
  
    Here, the environment variable should take the following format.
    
    ```
    <endpoint-name>_prod_endpoint_<endpoint-index>="<endpoint_URL>"
  
    <endpoint-name>  - Name specified in the swagger definition under x-wso2-endpoints.
    <endpoint-index> - The endpoint index starts from 0. Therefore, when overriding a single endpoint this value is 0.
    <endpoint_URL>   - URL of the new service endpoint.
    ``` 
    
    This will override the https://petstore.swagger.io/v2 endpoint you defined in the Dev environment with 
    https://petstore3.swagger.io/api/v3 when moving to the QA environment.
  
#### Changing Security

Security changes can be defined in the swagger definition and this definition can be used to deploy the API in the QA
environment. Operator will push a new image of the API to your docker registry in this case.

If you are using an existing image to deploy an API, there are some limitations when it comes to security as we need to
manually insert the certificate to the keystore of the Microgateway while changing the values in `micro-gw.conf` found in
the deployment. 
