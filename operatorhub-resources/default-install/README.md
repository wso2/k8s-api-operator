## How to install the API Operator 

- This documentation explains how to install the API Operator for kubernetes using Operator Lifecycle Manager
- In this example we shall use the same steps in [OperatorHub](https://operatorhub.io/operator/api-operator) to deploy the operator

Note:
- Here the API Operator will be created in **operators** namespace and will be usable from all namespaces in the cluster.
- However to ensure proper clean-up of resources, it is recommended to always deploy the API Operator in **wso2-namespace**
- To deploy the API Operator in aspecific namespace please follow [this](../namespace-install/README.md)

#### Prerequisites
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/)   
 - Minimum CPU and Memory for the K8s cluster: **2 vCPU, 8GB of Memory**

- An account in DockerHub or private docker registry
- [API controller](https://github.com/wso2/product-apim-tooling/releases/) 

#### Step 1: Install Operator Lifecycle manager
```
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/install.sh | bash -s 0.13.0
```
#### Step 2: Install the API Operator
```
kubectl create -f https://operatorhub.io/install/api-operator.yaml
```

When you execute the above command, it will create a Subscription in the "operators" namespace, including the details of the CSV source of the API operator.
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
- A subscription keeps CSVs up to date by tracking a channel in a package.
The above "my-api-operator" subscription is deployed in the "operators" namespace and we have provided the name and namespace of the catalog source coming with the operator lifecycle manager. We have also provided the operator name we want to deploy and the channel of the operator we want to subscribe.
- A CatalogSource is a repository of CSVs, CRDs, and packages that define an application. "operatorhubio-catalog" contains the CSVs, CRDs and packages of all the operators available in OperatorHub.io.
- A ClusterServiceVersion (CSV) is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator in a cluster. It contains the metadata such as name, version, icon, required resources, installation, etc.
- Once the above suscription is deployed, the required CRDs and CSV of the API Operator will be deployed in "operators" namespace, which will install the API Operator deployment with neccessary roles and bindings.

- You can check if the CSV has been properly deployed in operators namespace by executing the below command.
```
kubectl get csv -n wso2-system
```
- You can check if the operator is running by executing the below command
```
kubectl get pods -n wso2-system
```

### Using the API Operator

#### Step 1: Configure API Controller
- Download API controller v3.0.0 for your operating system from the [website](https://wso2.com/api-management/tooling/) or from [GitHub]((https://github.com/wso2/product-apim-tooling/releases/)
- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool
- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

You can find available operations using the below command.
```
apictl --help
```
By default API controller does not support kubectl command.
Set the API Controller’s mode to Kubernetes to be compatible with kubectl commands
```
apictl set --mode k8s 
```
#### Step 2: Install API Operator Configurations
* Execute the below command to create wso2-system namespace
```
kubectl create ns wso2-system
```
* Extract [api-k8s-crds-1.0.1.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.0.1/api-k8s-crds-1.0.1.zip). This zip contains the artifacts that required to deploy in Kubernetes. 
```
cd api-k8s-crds-1.0.1    
```
* Deploy the controller level configurations </br>
**[IMPORTANT]**   When you deploy an API, this will create a docker image for the API and be pushed to Docker-Hub. For this, your Docker-Hub credentials are required.   
1. Open **api-operator/controller-configs/controller_conf.yaml** and navigate to docker registry section(mentioned below), and  update ***user's docker registry***.       
```
#docker registry name which the mgw image to be pushed.
dockerRegistry: <username-docker-registry>        
``` 
2. Open **api-operator/controller-configs/docker_secret_template.yaml** and navigate to data section. </br> Enter the base 64 encoded username and password of the Docker-Hub account        
```        
data:         
  username: ENTER YOUR BASE64 ENCODED USERNAME         
  password: ENTER YOUR BASE64 ENCODED PASSWORD        
```        
Once you done with the above configurations, execute the following command to deploy controller configurations.       
```
kubectl apply -f api-operator/controller-configs/

configmap/controller-config created       
configmap/apim-config created   
security.wso2.com/default-security-jwt created       secret/wso2am310-secret created
configmap/docker-secret-mustache created        
secret/docker-secret created        
configmap/dockerfile-template created      
configmap/mgw-conf-mustache created
```

Now you can follow any scenarios provided [here](../../scenarios/README.md) and try out how the API Operator works.