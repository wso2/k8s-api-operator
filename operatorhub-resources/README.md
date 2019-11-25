## Instructions to test the operator bundle locally

#### Prerequisites
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- [Kubernetes v1.12 or above](https://Kubernetes.io/docs/setup/)   
 - Minimum CPU and Memory for the K8s cluster: **2 vCPU, 8GB of Memory**

- An account in DockerHub or private docker registry
- [API controller](https://github.com/wso2/product-apim-tooling/releases/) 


- Install OLM
```
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/olm.yaml
```

- Deploy operator-marketplace. (operator-marketplace operator is needed only for local testing)
```
git clone https://github.com/operator-framework/operator-marketplace.git
kubectl apply -f operator-marketplace/deploy/upstream/
```
- Deploy other resources
```
kubectl apply -f install.yaml 
```

**API operator will be deployed in kubernetes after some time (in the marketplace namespace for this example)**

### Checking the resources deployed

```
ramesha:workingdir ramesha$ kubectl get catalogsource -n marketplace
NAME                           NAME                           TYPE   PUBLISHER   AGE
rameshakaru-operators                                         grpc               18s
upstream-community-operators   Upstream Community Operators   grpc   Red Hat     43s
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ kubectl get opsrc rameshakaru-operators -o=custom-columns=NAME:.metadata.name,PACKAGES:.status.packages -n marketplace
NAME                    PACKAGES
rameshakaru-operators   apim-operator
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ kubectl get clusterserviceversion -n marketplace
NAME                   DISPLAY                       VERSION   REPLACES   PHASE
apim-operator.v1.0.1   API Operator for Kubernetes   1.0.1                Succeeded
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ kubectl get deployment -n marketplace
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
apim-operator                  1/1     1            1           58s
marketplace-operator           1/1     1            1           2m17s
rameshakaru-operators          1/1     1            1           75s
upstream-community-operators   1/1     1            1           100s
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ 
ramesha:workingdir ramesha$ kubectl get pods -n marketplace
NAME                                           READY   STATUS    RESTARTS   AGE
apim-operator-5db6d6cd67-zkqz4                 1/1     Running   0          73s
marketplace-operator-7cc57c5747-v2zgs          1/1     Running   0          2m32s
rameshakaru-operators-66b65df899-fwbs2         1/1     Running   0          90s
upstream-community-operators-5897c5d54-kqcwr   1/1     Running   0          115s

```
apim-operator is deployed

## How to preview the operator

Go to  https://operatorhub.io/preview and upload [csv file](apim-operator/1.0.1/apim-operator.v1.0.1.clusterserviceversion.yaml)

### Operator bundle

https://quay.io/application/rameshakaru/apim-operator?tab=releases
