### Kubernetes Extensions for API Operator

To work with operator, either [apictl tool](./configure-apictl-tool.md) or Kubernetes extensions are required.

Following extensions are provided as an alternative to work with API Operator.

You can add and update APIs with these extensions using a swagger file or an API project.

You can use API project to create APIs if you use interceptors. For more details about interceptors please refer scenario 10.

#### Install extensions

- Navigate to ***api-k8s-crds-1.0.0/apim-operator/kubectl-extensions*** directory.
- Make the extensions executable using the following command.
    ```$xslt
    chmod +x kubectl-*
    ```
- Copy the extensions into ***/usr/local/bin*** directory.
    ```$xslt
    cp kubectl-* /usr/local/bin/
    ```
    - **NOTE**: You MAY need to execute the COPY command with ***sudo***.
- Once the extensions are successfully configured execute the following commands to verify.
    
    - kubectl add Command
    ```$xslt
    kubectl add

    Error: must specify api, api name and Swagger file location.
    Add an API from a Swagger file.
    JSON and YAML formats are accepted.
    Examples:
      # Add an API using a Swagger file.
      kubectl add api petstore --from-file=./Swagger.json
      kubectl add api petstore --from-file=./Swagger.json --replicas=3
      kubectl add api petstore --from-file=./Swagger.json --namespace=wso2
      kubectl add api petstore --from-file=./Swagger.json --replicas=3 --namespace=wso2
      # Add an API using an API project
      kubectl add api petstore --from-file=petstore
      kubectl add api petstore --from-file=petstore --override=true
    Available Commands:
      api                 Add an API.
    Options:
      --filename='': Swagger file or location of the API project
      --replicas='': Number of replicas
    Usage:
      kubectl add api petstore --from-file=FILENAME [options] 
    
    ```
    - kubectl update command
    
    ```$xslt
    kubectl update

    Error: must specify api, api name and Swagger file location.
    Update an API from a Swagger file.
    JSON and YAML formats are accepted.
    Examples:
      # Update an API using a Swagger file.
      kubectl update api petstore --from-file=./Swagger.json
      kubectl update api petstore --from-file=./Swagger.json --replicas=3
      kubectl update api petstore --from-file=./Swagger.json --namespace=wso2
      kubectl update api petstore --from-file=./Swagger.json --replicas=3 --namespace=wso2
      # Update an API using an API project.
      kubectl update api petstore --from-file=petstore
    Available Commands:
      api                 Update an API.
    Options:
      --filename='': Swagger file or location of the API project
      --replicas='': Number of replicas
    Usage:
      kubectl update api petstore --from-file=FILENAME [options] 
    
    ```

#### Sample Usage

- Add APIs using the kubectl extensions

```$xslt
kubectl add api petapi --from-file=petstore-int --override=true

petstore-int is a directory

Creating interceptors configmap with name petapi-interceptors
configmap/petapi-interceptors created

Creating configmap with name petapi-swagger
configmap/petapi-swagger created
```

- Update APIs using kubectl extensions

```$xslt
Deleting interceptors configmap with name petapi-interceptors
configmap "petapi-interceptors" deleted

Creating interceptors configmap with name petapi-interceptors
configmap/petapi-interceptors created

Deleteting configmap if exists with name petapi-swagger
configmap "petapi-swagger" deleted

Creating configmap with name petapi-swagger
configmap/petapi-swagger created

Updating the API kind
api.wso2.com/petapi configured
```

