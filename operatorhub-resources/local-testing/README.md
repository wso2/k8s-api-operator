## Instructions to test the operator bundle locally

#### Prerequisites

- [Kubernetes v1.14 or above](https://Kubernetes.io/docs/setup/) <br>

    - Minimum CPU : 6vCPU
    - Minimum Memory : 6GB

- An account in DockerHub or private docker registry

- Download [k8s-api-operator-1.2.2.zip](https://github.com/wso2/k8s-api-operator/releases/download/v1.2.2/k8s-api-operator-1.2.2.zip) and extract the zip

    1. This zip contains the artifacts that required to deploy in Kubernetes.
    2. Extract k8s-api-operator-1.2.2.zip
    
    ```
    cd k8s-api-operator-1.2.2
    ```
 
    **_Note:_** You need to run all commands from within the ***k8s-api-operator-1.2.0*** directory.

<br />

- Install OLM
```shell script
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/olm.yaml
```

- Deploy operator-marketplace. (operator-marketplace operator is needed only for local testing)
```shell script
git clone https://github.com/operator-framework/operator-marketplace.git
kubectl apply -f operator-marketplace/deploy/upstream/
```
- Deploy other resources
```shell script
kubectl apply -f install.yaml 
```

**API operator will be deployed in kubernetes after some time (in the marketplace namespace for this example)**

#### Checking the resources deployed

```shell script
>> kubectl get catalogsource -n marketplace
NAME                           NAME                           TYPE   PUBLISHER   AGE
wso2am-operators                                              grpc               18s
upstream-community-operators   Upstream Community Operators   grpc   Red Hat     43s
 
>> kubectl get opsrc wso2am-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n marketplace
NAME                    PACKAGES
wso2am-operators   api-operator
 
>> kubectl get clusterserviceversion -n marketplace
NAME                   DISPLAY                       VERSION   REPLACES   PHASE
api-operator.v1.1.0   API Operator for Kubernetes     1.1.0              Succeeded

>> kubectl get deployment -n marketplace
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
api-operator                   1/1     1            1           58s
marketplace-operator           1/1     1            1           2m17s
wso2am-operators               1/1     1            1           75s
upstream-community-operators   1/1     1            1           100s

>> kubectl get pods -n marketplace
NAME                                           READY   STATUS    RESTARTS   AGE
api-operator-5db6d6cd67-zkqz4                  1/1     Running   0          73s
marketplace-operator-7cc57c5747-v2zgs          1/1     Running   0          2m32s
wso2am-operators-66b65df899-fwbs2              1/1     Running   0          90s
upstream-community-operators-5897c5d54-kqcwr   1/1     Running   0          115s

```
api-operator is deployed and is running in marketplace namespace.

#### How to preview the operator

Go to  https://operatorhub.io/preview and upload [csv file](api-operator/1.2.0/api-operator.v1.2.0.clusterserviceversion.yaml)

#### Operator bundle

https://quay.io/application/wso2am/api-operator?tab=releases