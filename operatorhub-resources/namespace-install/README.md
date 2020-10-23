## How to install the API Operator in a specific namespace

- This documentation explains how to install the API Operator for kubernetes in a specific namespace using Operator Lifecycle Manager
- In this example we shall deploy the operator in **wso2-system** namespace

Note:
- To ensure proper clean-up of resources, it is recommended to always deploy the API Operator in **wso2-namespace**
- However, the API Operator is capable of functioning from any namespace

#### Prerequisites
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.14 or above](https://Kubernetes.io/docs/setup/)   
 - Minimum CPU and Memory for the K8s cluster: **2 vCPU, 8GB of Memory**

- An account in DockerHub or private docker registry
- [API controller](https://github.com/wso2/product-apim-tooling/releases/) 

#### Step 1: Install Operator Lifecycle manager
```
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/install.sh | bash -s 0.13.0
```
#### Step 2: Install the API Operator
```
kubectl apply -f install.yaml
```

When you execute the above command, the following will happen.

- Create a namespace of wso2-system
- Create an OperatorGroup in wso2-system namespace giving access to wso2-system namespace for its members. (i.e. to deploy the cluster service version of the operator in wso2-system namespace.) 
```
apiVersion: operators.coreos.com/v1alpha2
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: wso2-system
spec:
  targetNamespaces:
  - wso2-system
```

An OperatorGroup selects a set of target namespaces in which to generate required RBAC access for its member operators and there can be only one OperatorGroup in one namespace. If not, the CSV deployment will fail with "TooManyOperatorGroups" reason.

- Create a Subscription in the same namespace as the OperatorGroup, including the details of the CSV source of the API operator.
```
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: my-api-operator
  namespace: wso2-system
spec:
  channel: stable
  name: api-operator
  source: operatorhubio-catalog
  sourceNamespace: olm
```
A subscription keeps CSVs up to date by tracking a channel in a package.
The above "my-api-operator" subscription is deployed in the wso2-system namespace and we have provided the name and namespace of the catalog source coming with the operator lifecycle manager. We have also provided the operator name we want to deploy and the channel of the operator we want to subscribe.
</br>
A CatalogSource is a repository of CSVs, CRDs, and packages that define an application. "operatorhubio-catalog" contains the CSVs, CRDs and packages of all the operators available in OperatorHub.io.
</br>
A ClusterServiceVersion (CSV) is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator in a cluster. It contains the metadata such as name, version, icon, required resources, installation, etc.
</br>
Once the above suscription is deployed, the required CRDs and CSV of the API Operator will be deployed in wso2-system namespace, which will install the API Operator deployment with neccessary roles and bindings.

- You can check if the CSV has been properly deployed in wso2-system namespace by executing the below command.
```
kubectl get csv -n wso2-system
```
- You can check if the operator is running by executing the below command
```
kubectl get pods -n wso2-system
```

### Using the API Operator

#### Step 1: Configure API Controller
- Download API controller v3.1.0 or the latest v3.1.x from the [API Manager Tooling web site](https://wso2.com/api-management/tooling/)
    - Under Dev-Ops Tooling section, you can download the tool based on your operating system.
- Extract the API controller distribution and navigate inside the extracted folder using the command-line tool
- Add the location of the extracted folder to your system's $PATH variable to be able to access the executable from anywhere.

You can find available operations using the below command.
```
>> apictl --help
```
By default API controller does not support kubectl command.
Set the API Controller’s mode to Kubernetes to be compatible with kubectl commands
```
>> apictl set --mode k8s 
```

#### Step 2: Install API Operator Configurations

* Download [k8s-api-operator-1.2.2.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.2/api-k8s-crds-1.2.2.zip)
  1. This zip contains the artifacts that required to deploy in Kubernetes.
  2. Extract k8s-api-operator-1.2.2.zip
```
cd k8s-api-operator-1.2.2
```     
* Create namspace and deploy the controller level configurations **[IMPORTANT]**
  *  When you create an API, a docker image of it will be created and pushed to a docker registry. For this, credentials for your docker resgitry are required.
  
```
>>  apictl change registry
    
Choose registry type:
1: Docker Hub (Or others, quay.io, HTTPS registry)
2: Amazon ECR
3: GCR
4: HTTP Private Registry
Choose a number: 1: 1
Enter repository name (docker.io/john | quay.io/mark | 10.100.5.225:5000/jennifer): docker.io/jennifer
Enter username: jennifer
Enter password: *******
Repository: docker.io/jennifer
Username  : jennifer
Confirm configurations: Y: Y

Output:

configmap/docker-registry-config configured
secret/docker-registry-credentials configured
```
Once you are done with the above configurations, execute the following command to deploy controller configurations.

```
>> kubectl apply -f api-operator/operatorhub-controller-configs/

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

Now you can follow any scenarios provided [here](../../scenarios/README.md) and try out how the API Operator works.