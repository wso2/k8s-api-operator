### Kubernetes Extensions for API Operator

To work with operator, either [apictl tool](./configure-apictl-tool.md) or Kubernetes extensions are required.

Following extensions are provided as an alternative to work with API Operator.

#### Install extensions

- Navigate to ***api-k8s-crds-1.0.0/apim-operator/kubectl-extensions*** directory.
- Make the extensions executable using the following command.
    ```$xslt
    chmod +x ./deploy/kubectl-extension/kubectl-add
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
    Available Commands:
      api                 Add an API.
    Options:
      --filename='': File Location
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
      # Add an API using a Swagger file.
      kubectl update api petstore --from-file=./Swagger.json
      kubectl update api petstore --from-file=./Swagger.json --replicas=3
      kubectl update api petstore --from-file=./Swagger.json --namespace=wso2
      kubectl update api petstore --from-file=./Swagger.json --replicas=3 --namespace=wso2
    Available Commands:
      api                 Update an API.
    Options:
      --filename='': File Location
      --replicas='': Number of replicas
    Usage:
      kubectl update api petstore --from-file=FILENAME [options] 
    ```

