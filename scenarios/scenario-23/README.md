## Scenario 23 - Add Config Maps and Secrets to Micro-gateway Deployment

- This scenario describes how to add or mount the desired config maps and secrets to micro-gateway deployment.

 ***Important:***
> Follow the main README and deploy the api-operator and configuration files. Make sure to set the analyticsEnabled to "true" and deploy analytics secret with credentials to analytics server and certificate, if you want to check analytics.

##### Deploying the artifacts 

- Create a namespace as "micro".
  ```
  >> apictl create ns micro
      
  Output:
  namespace/micro created    
  ```
- Navigate to scenarios/scenario-23 directory.

- Create and deploy the config maps and secrets that you need to mount. In here
`test1cm` config map is in "wso2-system" namespace. `test2cm` config map and `test1secret` 
secret are in "micro" namespace.
  ```$xslt
  >> apictl apply -f test1cm.yaml
  >> apictl apply -f test2cm.yaml
  >> apictl apply -f test1secret.yaml     
   
  Output:
  configmap/test1cm created
  configmap/test2cm created
  secret/test1secret created
  ```
  
#### Approach 01

- Micro-gateway Deployment Configurations can be configured in the mgw-deployment-configs in `controller-configs/controller_conf.yaml`

     ```
      mgwConfigMaps: |
        - name: test1cm
          mountLocation: /home/ballerina/test1cm
          subPath: test1cm
      mgwSecrets: |
        - name: test1secret
          mountLocation: /home/ballerina/test1secret
          subPath: test1secret
          namespace: micro
     ```
- In this configuration you can provide: 
    - name of the config map or secret.
    - location that you need to mount the config map or secret.
    - sub path of the mount location. 
    - namespace of the config map or secret. This field is an optional field.
    When namespace is not provided, the config map or secret should be created in the same namespace
    which the mgw-deployment-configs are deployed. In this case `test1cm` is deployed in
    "wso2-system" namespace. 

***Important:***
> If you create an API in another namespace (Eg: foo) then, only config maps or secrets which a namespace
>is not specified under mgw-deployment-configs (Eg: test1cm) will be mounted. 

- Apply the changes
    ```$xslt
    >> apictl apply -f <k8s-api-operator-home>/api-operator/controller-artifacts/controller_conf.yaml
    ```
  
- Execute the following to expose pet-store service as an API.

- Create API <br /> 
    ```
    >> apictl add api -n petstore-api --from-file=swagger.yaml --override --namespace=micro
        
    Output:
    creating configmap with swagger definition
    configmap/petstore-api-swagger created
    creating API definition
    api.wso2.com/petstore-api created
    ``` 
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.

- Get the pods 
    ```
    >> apictl get pods -n micro
            
    Output:
    NAME                            READY   STATUS      RESTARTS   AGE
    petstore-api-5fcf4d4cc8-fbczp   1/1     Running     0          7m42s
    ```

- Describe the pods. You can see the config map and the secret has been mounted.
    ```
    Mounts:
       /home/ballerina/test1cm from test1cm-vol (rw,path="test1cm")
       /home/ballerina/test1secret from test1secret-vol (rw,path="test1secret")
    ```

- Delete the  API
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
        ```
        >> apictl delete api petstore-api -n micro
        
        Output:
        api.wso2.com "petstore-api" deleted
        ```
      
#### Approach 02

- You can create and deploy `mgw-deployment-configs` config map in the same namespace that the API is being deployed 
without altering the mgw-deployment-configs in `controller-configs/controller_conf.yaml`
    ```  
    >> apictl apply -f mgw-deploy-configs.yaml
    
    Output:
    configmap/mgw-deployment-configs created
    ```

- In this configuration you can provide: 
    - name of the config map or secret.
    - location that you need to mount the config map or secret.
    - sub path of the mount location. 
    - namespace of the config map or secret. This field is an optional field.
    When namespace is not provided, the config map or secret should be created in the namespace
    which the mgw-deployment-configs are deployed. In this case `test1secret` secret is deployed in
    "micro" namespace.
    
- Execute the following to expose pet-store service as an API.

- Create API <br /> 
    ```
    >> apictl add api -n petstore-api --from-file=swagger.yaml --override --namespace=micro
        
    Output:
    creating configmap with swagger definition
    configmap/petstore-api-swagger created
    creating API definition
    api.wso2.com/petstore-api created
    ``` 
    Note: ***--override*** flag is used to you want to rebuild the API image even if it exists in the configured docker repository.

- Get the pods 
    ```
    >> apictl get pods -n micro
            
    Output:
    NAME                            READY   STATUS      RESTARTS   AGE
    petstore-api-5fcf4d4cc8-fbczp   1/1     Running     0          7m42s
    ```

- Describe the pods. You can see the config maps and the secrets have been mounted.
    ```
    Mounts:
       /home/ballerina/test1secret from test1secret-vol (rw,path="test1secret")
       /home/ballerina/test2cm from test2cm-vol (rw,path="test2cm")
    ```

- Delete the  API
    - Following command will delete all the artifacts created with this API including pods, deployment and services.
        ```
        >> apictl delete api petstore-api -n micro
        
        Output:
        api.wso2.com "petstore-api" deleted
        ```

- Delete config maps and secrets.
    ```
    >> apictl delete ns micro
    >> apictl delete configmap test1cm -n wso2-system
               
    Output:
    namespace "micro" deleted
    configmap "test1cm" deleted
    ```
