##Trouble Shooting Guide - API Operator

#### Check logs in API Operator
- API operator deploys in a namespace called "wso2-system"

- Following command will list the available pods in the "wso2-system" namespace.

    ```$xslt
    kubectl get pods -n wso2-system
    ``` 
- Output:
    ```$xslt
    NAME                             READY   STATUS    RESTARTS   AGE
    apim-operator-59c665f477-9bw7l   1/1     Running   0          4h23m
     
    ```
- Once you are able to see the apim-operator pod up and running, you can check its logs using the below command.
    
    ```$xslt
    kubectl logs -f -n wso2-system <name of the apim-operator pod>
    ```
- Example: 

    ```$xslt
    kubectl logs -f -n wso2-system apim-operator-59c665f477-9bw7l
    ```
- Once the above command is executed, it will show the logs in the API operator.

#### Identifying Kaniko job related pod & errors

- Kaniko job is responsible to create the API microgateway image and push it to the Docker-Hub.
- If the API microgateway image belongs to a particular API definition is not available in the Docker-Hub, it will build the image using the Kaniko job.
- If you are creating an API name "online-store", the Kaniko pod related to that would look like below.
    <api-name>-kaniko-xxxxxx-xxxx
      
    online-store-kaniko-xxxxxx-xxxx (x denotes random alphanumeric values)

```$xslt
kubectl get pods
```

```$xslt
NAME                                   READY   STATUS    RESTARTS   AGE    
online-storee-kaniko-6dvb8             1/1     Running   0          5s

```
- If it's in the running "status", it's working fine. If it says "Err", most possibly it can be due to configuration issue related to Docker-Hub user. Hence pushing the image may leads the kaniko pod to a erroneous state.
- In that case check the following,
    1.Check the if you have put the proper Docker-Hub username in "<api-k8s-crd-home>/apim-operator/controller-configs/controller_conf.yaml" 
    - Check the following configuration in <api-k8s-crd-home>/apim-operator/controller-configs/controller_conf.yaml.
    - Replace the <username-docker-registry> with the proper Docker-Hub username.   
        ```
        #docker registry name which the mgw image to be pushed.  eg->  dockerRegistry: username
        dockerRegistry: <username-docker-registry>
        ```  
        ```$xslt
        kubectl apply -f <api-k8s-crd-home>/apim-operator/controller-configs/controller_conf.yaml
        ```
    - Once it's modified, execute the following command to apply the changes in the cluster
    2.Check if you have provided the Docker-Hub username and password in the docker_secret_template.file.
    - Open the <api-k8s-crd-home>/apim-operator/controller-configs/docker_secret_template.yaml file. 
    - Check if you have entered the **base 64 encoded value of username and password** in the following section.
        ```$xslt
        data:
          username: ENTER YOUR BASE64 ENCODED USERNAME
          password: ENTER YOUR BASE64 ENCODED PASSWORD
        ``` 
        ```$xslt
        kubectl apply -f <api-k8s-crd-home>/apim-operator/controller-configs/docker_secret_template.yaml
        ```
#### How to check logs in API

- Once the API is deploy in the Kubernetes cluster, the pod will be names in the following convention.
  <api-name>-xxxxx-xxxx (x is a alphanumeric value)
- If you have deployed online-store API, the pod will be look like below.
    ```$xslt
    NAME                                   READY   STATUS      RESTARTS   AGE
    online-store-794cd7b66-lnnxd           1/1     Running     0          164m
    ```
- To check it's log, execute the following command.
    ```$xslt
    kubectl logs -f online-store-794cd7b66-lnnxd
    ```   
    Sample logs of the API as below.
    ```$xslt
    [ballerina/http] started HTTPS/WSS endpoint 0.0.0.0:9096
    [ballerina/http] started HTTPS/WSS endpoint 0.0.0.0:9095
    [ballerina/http] started HTTP/WS endpoint 0.0.0.0:9090
    2019-10-27 14:37:49,222 INFO  [wso2/gateway] - HTTPS listener is active on port 9095 
    2019-10-27 14:37:49,224 INFO  [wso2/gateway] - HTTP listener is active on port 9090
    ```

#### How to enable debug logs for the API

- If you want to analyze logs in depth, enable the debug logs.
- For this, you need to add the following entry in the ***<api-k8s-crds-home>/apim-operator/controller-configs/mgw_conf_mustache.yaml***
```$xslt
[b7a.log]
level="DEBUG"
```
- Reapply this configuration separately using the following command.
```$xslt
kubectl apply -f <api-k8s-crds-home>/apim-operator/controller-configs/mgw_conf_mustache.yaml
```
- Once you apply this, you need to build the API from scratch to reflect these changes to the already deployed APIs.

