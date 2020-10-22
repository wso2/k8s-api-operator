## How to install the API Operator 

- This documentation explains how to install the API Operator for kubernetes using Operator Lifecycle Manager
- In this example we shall use the same steps in [OperatorHub](https://operatorhub.io/operator/api-operator) to deploy the operator

Note:
- Here the API Operator will be created in **operators** namespace and will be usable from all namespaces in the cluster.
- However to ensure proper clean-up of resources, it is recommended to always deploy the API Operator in **wso2-namespace**.
- To deploy the API Operator in aspecific namespace please follow [this](../namespace-install/README.md).

#### Prerequisites

- [Kubernetes v1.14 or above](https://Kubernetes.io/docs/setup/) <br>

    - Minimum CPU : 6vCPU
    - Minimum Memory : 6GB

- An account in DockerHub or private docker registry

- Download [k8s-api-operator-1.2.0.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.0/k8s-api-operator-1.2.0.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract k8s-api-operator-1.2.0.zip
    
    ```
    cd k8s-api-operator-1.2.2
    ```
 
    **_Note:_** You need to run all commands from within the ***k8s-api-operator-1.2.0*** directory.

<br />

#### Step 1: Install Operator Lifecycle manager

```shell script
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/install.sh | bash -s 0.13.0
```
#### Step 2: Install the API Operator

```shell script
>> kubectl create -f https://operatorhub.io/install/api-operator.yaml
```

- When you execute the above command, it will create a Subscription in the "operators" namespace, including the details of the CSV source of the API operator.
  
    ```
    apiVersion: operators.coreos.com/v1alpha1
    kind: Subscription
    metadata:
      name: my-api-operator
      namespace: operators
    spec:
      channel: stable
      name: api-operator
      source: operatorhubio-catalog
      sourceNamespace: olm
    ```
- A subscription keeps CSVs up to date by tracking a channel in a package. The above "my-api-operator" subscription is deployed in the "operators" namespace and we have provided the name and namespace of the catalog source coming with the operator lifecycle manager. We have also provided the operator name we want to deploy and the channel of the operator we want to subscribe.

- A CatalogSource is a repository of CSVs, CRDs, and packages that define an application. "operatorhubio-catalog" contains the CSVs, CRDs and packages of all the operators available in OperatorHub.io.

- A ClusterServiceVersion (CSV) is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator in a cluster. It contains the metadata such as name, version, icon, required resources, installation, etc.

- Once the above subscription is deployed, the required CRDs and CSV of the API Operator will be deployed in "operators" namespace, which will install the API Operator deployment with neccessary roles and bindings.

- You can check if the CSV has been properly deployed in operators namespace by executing the below command.
    ```shell script
    >> kubectl get csv -n wso2-system
    ```
- You can check if the operator is running by executing the below command
    ```shell script
    >> kubectl get pods -n wso2-system
    ```

### Using the API Operator

#### Step 1: Configure API Controller

- Download API controller v3.1.0 or the latest v3.1.x from the [API Manager Tooling web site](https://wso2.com/api-management/tooling/)
    
    - Under Dev-Ops Tooling section, you can download the tool based on your operating system.

- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool

- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

- You can find available operations using the below command.
    
    ```shell script
    >> apictl --help
    ```
  
- By default API controller does not support kubectl command.

- Set the API Controller’s mode to Kubernetes to be compatible with kubectl commands
    ```shell script
    >> apictl set --mode k8s 
    ```

#### Step 2: Install API Operator Configurations

* Download [k8s-api-operator-1.2.0.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.0/api-k8s-crds-1.2.0.zip)
  1. This zip contains the artifacts that required to deploy in Kubernetes.
  2. Extract k8s-api-operator-1.2.0.zip
  
    ```shell script
    cd k8s-api-operator-1.2.2
    ```     
* Create a namespace and deploy the controller level configurations 

    ```shell script
    >> apictl apply -f api-operator/operatorhub-controller-configs/
    
    namespace/wso2-system created
    configmap/controller-config created
    configmap/apim-config created
    configmap/ingress-configs created
    configmap/kaniko-arguments created
    configmap/route-configs created
    security.wso2.com/default-security-jwt created
    secret/wso2am320-secret created
    configmap/docker-registry-config ceated
    configmap/dockerfile-template created
    configmap/mgw-conf-mustache unchanged
    ```
  
*  When you create an API, a docker image of it will be created and pushed to a docker registry. For this, credentials for your docker registry are required.
  
    ```shell script
    >>  apictl change registry
        
    Choose registry type:
    1: Docker Hub
    2: Amazon ECR
    3: GCR
    4: HTTP Private Registry
    5: HTTPS Private Registry
    6: Quay.io
    Choose a number: 1: 1
    Enter repository name: docker.io/jennifer
    Enter username: jennifer
    Enter password: *******
    Repository: docker.io/jennifer
    Username  : jennifer
    Confirm configurations: Y: Y
    
    Output:
    
    configmap/docker-registry-config configured
    secret/docker-registry-credentials configured
    ```
   
* Now you can follow any scenarios provided [here](../../scenarios/README.md) and try out how the API Operator works.