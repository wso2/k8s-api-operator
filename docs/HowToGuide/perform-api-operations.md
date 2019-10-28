## Perform API Operations

This guide explains how you can add, update and delete an API. Also operations you can carried out for the API custom resource.

#### Add an API

Using the following command, you can add an API to the microservice.

```
Format:
>> apictl add api -n "api_name" --from-file="location to the Open API definition"
```

```
>> apictl add api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml

Output:
creating configmap with swagger definition
configmap/online-store-swagger created
api.wso2.com/online-store created
```

Optional Parameters

```
--replicas=3          Number of replicas
--namespace=wso2      Namespace to deploy the API
--overwrite=true	  Overwrite the docker image creation for already created docker image

>> apictl add api -n "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" --namespace="desired namespace"
```

#### Update an API

Using the following command you can update the deployed API.

```
Format:
>> apimcli update api -n "api_name" --from-file="location to the api swagger definition"

>> apimcli update api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml

Output:
creating configmap with swagger definition
configmap/online-store-swagger-up-20191028143610 created
updating the API Kind
api.wso2.com/online-store configured
```

When you update the API, it will create a new docker image and push to the docker registry. Then it will perform a rolling update to deploy the new API gateway.

#### Delete an API

Using the following command you can delete the deployed API and the artifacts related to that API (pods, deployment, service, HPA)

```
Format:
>> apictl delete api "api_name"

>> apictl delete api online-store

Output:
api.wso2.com "online-store" deleted
```

#### List the APIs in Kubernetes

You can list down the APIs as follows.
```
>> kubectl get apis

Output:
NAME       AGE
online-store   20h
```

#### Describe the API in Kubernetes


```
>> kubectl describe api online-store

Output:
Name:         online-store
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"wso2.com/v1alpha1","kind":"API","metadata":{"annotations":{},"creationTimestamp":null,"name":"online-store","namespace":"defaul...
API Version:  wso2.com/v1alpha1
Kind:         API
Metadata:
  Creation Timestamp:  2019-10-27T12:18:00Z
  Generation:          2
  Resource Version:    585096
  Self Link:           /apis/wso2.com/v1alpha1/namespaces/default/apis/online-store
  UID:                 d107e14d-f8b3-11e9-9a6a-42010a8001fc
Spec:
  Definition:
    Configmap Name:       online-store-swagger-up-20191028143610
    Type:                 swagger
  Interceptor Conf Name:
  Mode:                   privateJet
  Override:               false
  Replicas:               1
  Update Time Stamp:      20191028143610
Status:
Events:  <none>
```
