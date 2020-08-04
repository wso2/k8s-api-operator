### Perform API Operations

This guide explains how you can add, update and delete an API. Also operations you can carried out for the API custom resource.

#### Add an API

Using the following command, you can add an API to the microservice.

```sh
Format:
>> apictl add api -n <API_NAME> --from-file=<PATH_TO_OPEN_API_DEFINITION>
>> apictl add api -n <API_NAME> --from-file=<PATH_TO_API_PROJECT>
```
***NOTE:*** Here ***API_PROJECT*** refers to a project initialized using `apictl init command`

```sh
>> apictl add api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml

Output:
Processing swagger 1: scenarios/scenario-1/products_swagger.yaml
creating configmap with swagger definition
configmap/online-store-1-swagger created
creating API definition
api.wso2.com/online-store created
```

You can also provide multiple open API definitions for a API

```sh
Format:
>> apictl add api -n <API_NAME> \
        --from-file=<PATH_TO_OPEN_API_DEFINITION_1> \
        --from-file=<PATH_TO_OPEN_API_DEFINITION_2> \
        ... --from-file=<PATH_TO_OPEN_API_DEFINITION_m> \
        --from-file=<PATH_TO_API_PROJECT_1> \
        --from-file=<PATH_TO_API_PROJECT_2> \
        ... --from-file=<PATH_TO_API_PROJECT_n>
```

Optional Parameters

```sh
--replicas=3          Number of replicas
--namespace=wso2      Namespace to deploy the API
--override      	  Overwrite the docker image creation for already created docker image
--mode                Overwrite the deploying mode when multiple open API definitions provided. Available modes: privateJet, sidecar
--version             Overwrite the deploying version when multiple open API definitions provided
--env                 Environment variables to be passed to deployment
--image               Image of the API. If specified, ignores the value of --override

Format:
>> apictl add api -n <API_NAME> --from-file=<PATH_TO_OPEN_API_DEFINITION_1> --from-file=<PATH_TO_API_INIT_PROJECT_1> ... \
        --replicas=<NUMBER_OF_REPLICAS> \
        --namespace=<DESIRED_NAMESPACE> \
        --mode=<DEPLOY_MODE> \
        --version=<DEPLOY_VERSION> \
        --env=<KEY_1>=<VALUE_1> --env=<KEY_2>=<VALUE_2> \
        --image=<EXISTING_IMAGE>
```

***NOTE:*** Flags `--mode` and `--version` only supports in multiple open API definitions mode. If these values are not provided default mode: `privateJet` and default version: `v1.0.0.` are used.

#### Update an API

Using the following command you can update the deployed API. As in the add API you can use multiple open API definitions.

```sh
Format:
>> apictl update api -n <API_NAME> --from-file=<PATH_TO_OPEN_API_DEFINITION_1>
```

```sh
>> apictl update api -n online-store --from-file=scenarios/scenario-1/products_swagger.yaml

Output:
Processing swagger 1: scenarios/scenario-1/products_swagger.yaml
creating configmap with swagger definition
configmap/online-store-1-swagger created
creating API definition
api.wso2.com/online-store configured
```

When you update the API, it will create a new docker image and push to the docker registry. Then it will perform a rolling update to deploy the new API gateway.

#### Delete an API

Using the following command you can delete the deployed API and the artifacts related to that API (pods, deployment, service, HPA)

```sh
Format:
>> apictl delete api <API_NAME>

>> apictl delete api online-store

Output:
api.wso2.com "online-store" deleted
```

#### List the APIs in Kubernetes

You can list down the APIs as follows.
```sh
>> apictl get apis

Output:
NAME       AGE
online-store   20h
```

#### Describe the API in Kubernetes

```sh
>> apictl describe api online-store

Output:
Name:         online-store
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"wso2.com/v1alpha1","kind":"API","metadata":{"annotations":{},"creationTimestamp":null,"name":"online-store","namespace":"de...
API Version:  wso2.com/v1alpha1
Kind:         API
Metadata:
  Creation Timestamp:  2020-04-04T01:12:57Z
  Generation:          1
  Resource Version:    13839
  Self Link:           /apis/wso2.com/v1alpha1/namespaces/default/apis/online-store
  UID:                 6b2775ae-7611-11ea-8395-42010a8000a7
Spec:
  Definition:
    Interceptors:
    Swagger Configmap Names:
      online-store-1-swagger
    Type:    swagger
  Mode:      privateJet
  Override:  true
  Replicas:  1
Events:      <none>
```
