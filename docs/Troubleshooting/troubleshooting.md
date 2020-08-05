## Trouble Shooting Guide - API Operator

#### Check logs in API Operator
- API operator deploys in a namespace called "wso2-system"

- Following command will list the available pods in the "wso2-system" namespace.

    ```sh
    >> apictl get pods -n wso2-system
  
    Output:
    NAME                             READY   STATUS    RESTARTS   AGE
    api-operator-59c665f477-9bw7l   1/1     Running   0          4h23m
     
    ```
- Once you are able to see the api-operator pod up and running, you can check its logs using the below command.
    
    ```sh
    >> apictl logs -f -n wso2-system <NAME_OF_THE_API-OPERATOR_POD>
    ```
- Example: 

    ```sh
    >> apictl logs -f -n wso2-system api-operator-59c665f477-9bw7l
    ```
- Once the above command is executed, it will show the logs in the API operator.

#### Identifying Kaniko job related pod and errors

- Kaniko job is responsible to create the API Microgateway image and push it to the registry configured during the
  API operator installation.
- If the API Microgateway image belongs to a particular API definition is not available in the docker repository,
  it will build the image using the Kaniko job.
- If you are creating an API name "online-store", the Kaniko pod related to that would look like below. <br>
    `<API_NAME>-kaniko-xxxxxx-xxxx`
- Example:
      
    `online-store-kaniko-xxxxxx-xxxx` (x denotes random alphanumeric values)

    ```sh
    >> apictl get pods
    
    Output:
    NAME                                   READY   STATUS    RESTARTS   AGE    
    online-storee-kaniko-6dvb8             1/1     Running   0          5s
    ```

- If it's in the running "status", it's working fine. If it says "Err", most possibly it can be due to configuration
  issue related to docker registry credentials or connection to the registry. Hence pushing the image may leads
  the kaniko pod to a erroneous state.

- Find the logs in the Kaniko job for more information and get description about the pod that runs the Kaniko job.
    ```sh
    >> apictl describe pod <POD_NAME_OF_KANIKO_JOB>
    >> apictl logs -f <POD_NAME_OF_KANIKO_JOB>
    ```

- If the error is related to authentication, reconfigure registry credentials using `apictl` tool. Go through the
  interactive session to reconfigure credentials.
    ```sh
    >> apictl change registry
    ```

#### How to check logs in API

- Once the API is deploy in the Kubernetes cluster, the pod will be names in the following convention.

    `<api-name>-xxxxx-xxxx` (x is a alphanumeric value)
    
- If you have deployed online-store API, the pod will be look like below.
    ```sh
    NAME                                   READY   STATUS      RESTARTS   AGE
    online-store-794cd7b66-lnnxd           1/1     Running     0          164m
    ```
  
- To check its log, execute the following command.
    ```sh
    >> apictl logs -f online-store-794cd7b66-lnnxd
    
    Sample logs:
    [ballerina/http] started HTTPS/WSS endpoint 0.0.0.0:9096
    [ballerina/http] started HTTPS/WSS endpoint 0.0.0.0:9095
    [ballerina/http] started HTTP/WS endpoint 0.0.0.0:9090
    2019-10-27 14:37:49,222 INFO  [wso2/gateway] - HTTPS listener is active on port 9095 
    2019-10-27 14:37:49,224 INFO  [wso2/gateway] - HTTP listener is active on port 9090
    ```

#### How to enable debug logs for the API

- If you want to analyze logs in depth, enable the debug logs.
- For this, you need to update the `logLevel` field of the configmap: `apim-config` in the file
  ***\<k8s-api-operator-home>/api-operator/deploy/controller-configs/controller_conf.yaml*** to "DEBUG".
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: apim-config
      namespace: wso2-system
    data:
      #Log level of the managed API (microgateway). Available levels: INFO, DEBUG, TRACE
      logLevel: "DEBUG"
      ...
    ```

- Reapply this configuration separately using the following command.
    ```sh
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-configs/controller_conf.yaml
    ```
- Once you apply this, you need to build the API from scratch to reflect these changes to the already deployed APIs.

