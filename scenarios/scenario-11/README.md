## Scenario 11 - Enabling Analytics for managed API


- This scenario describes how to enable analytics in the api-operator and monitor analytics in the analytics dashboard

This setup provides resources to deploy WSO2 API Manager 3.1.0 and WSO2 APIM Analytics 3.1.0 in the Kubernetes cluster and configure them with WSO2 Microgateway using k8s CRD operator.
 
Few databases of WSO2 API Manager and Analytics need to be shared with each other and hence we have provided resources to deploy mysql in kubernetes and configure the necessary databases in mysql.


To try out the scenario navigate to ```k8s-api-operator-<version>``` directory.

#### Step 1: Deploy and configure API Portal and Analytics Dashbaord

[WSO2AM Kubernetes Operator](https://github.com/wso2/k8s-wso2am-operator) is used to deploy API portal and Analytics dashboard. 

- Install the WSO2AM Operator in Kubernetes.

    ```
    >> apictl install wso2am-operator
    
    namespace/wso2-system configured
    serviceaccount/wso2am-pattern-1-svc-account created
    ...
    configmap/wso2am-p1-apim-2-conf created
    configmap/wso2am-p1-mysql-dbscripts created
    [Setting to K8s Mode]
    ```

- Install API Portal and API Analytics in a namespace called "wso2"

    ```
    >> apictl apply -f k8s-artifacts/wso2am-operator/api-portal-with-analytics/wso2-namespace.yaml
    
    Output:
    namespace/wso2 created

    >> apictl apply -f k8s-artifacts/wso2am-operator/api-portal-with-analytics/mysql/
    
    Output:
    configmap/mysql-dbscripts created
    deployment.apps/wso2apim-with-analytics-mysql-deployment created
    service/wso2apim-with-analytics-rdbms-service created
    
    >> apictl apply -f k8s-artifacts/wso2am-operator/api-portal-with-analytics/configmaps/
    
    Output:
    configmap/dash-conf created
    configmap/worker-conf created
    configmap/apim-conf created
    
    >> apictl apply -f k8s-artifacts/wso2am-operator/api-portal-with-analytics/custom-pattern.yaml
    
    Output:
    apimanager.apim.wso2.com/custom-pattern-2 created
    ```

- Access API Portal and API Analytics 

    ```
    >> apictl get pods -n wso2
    
    Output:
    NAME                                                        READY   STATUS    RESTARTS   AGE
    all-in-one-api-manager-5694f99754-4zhpq                     1/1     Running   0          2m39s
    analytics-dashboard-6d6c5dd-2xd4r                           1/1     Running   0          2m39s
    analytics-worker-6c64b9bd79-794l7                           1/1     Running   0          2m39s
    wso2apim-with-analytics-mysql-deployment-6659655c65-njs7r   1/1     Running   0          3m3s
    
    >> apictl get services -n wso2
    
    Output:
    NAME                                    TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                                                       AGE
    wso2apim                                NodePort    10.0.38.209   <none>        8280:32004/TCP,8243:32003/TCP,9763:32002/TCP,9443:32001/TCP   3m43s
    wso2apim-analytics-service              ClusterIP   10.0.47.66    <none>        7612/TCP,7712/TCP,9444/TCP,9091/TCP,7071/TCP,7444/TCP         3m44s
    wso2apim-dashboard-analytics-service    NodePort    10.0.35.34    <none>        9643:32201/TCP                                                3m44s
    wso2apim-with-analytics-rdbms-service   ClusterIP   10.0.36.229   <none>        3306/TCP                                                      4m7s
    ```
    
    **Note:** To access the API portal and Analytics dashboard, add host mapping entries to the /etc/hosts file. As we have exposed the services in Node Port type, you can use the IP address of any Kubernetes node.
    
    ```
    <ANY_K8S_NODE_IP>  wso2apim
    <ANY_K8S_NODE_IP>  wso2apim-analytics
    ```

    - For Docker for Mac use "127.0.0.1" for the K8s node IP
    - For Minikube, use minikube ip command to get the K8s node IP
    - For GKE
        ```$xslt
        (apictl get nodes -o jsonpath='{ $.items[*].status.addresses[?(@.type=="ExternalIP")].address }')
        ```
        - This will give the external IPs of the nodes available in the cluster. Pick any IP to include in /etc/hosts file.
      
       **API Portal** - https://wso2apim:32001/devportal <br>
       **API Analytics Dashbaord** - https://wso2apim-analytics:32201/analytics-dashboard


#### Step 2: Enable API Analytics in the API Operator

- If you haven't deployed the API Operator please follow the quick start guide in root readme and follow step 2 and 3.
- By deploying the analytics configmaps, you can enable analytics as follows.

    ```
    >> apictl apply -f api-operator/apim-analytics-configs
    
    ---
    configmap/analytics-config created
    secret/analytics-secret created
    secret/wso2analytics300-secret created
    ---
    ```

#### Step 3: Deploy an API

- Please follow the scenario 1 readme and deploy the API.

- Use the API Analytics dashboard url which configured in Step 1 to browse the analytics information for the API.

- You will be able to monitor the analytics as shown in below images.

##### [Analytics Dashboard](https://wso2apim-analytics:32201/analytics-dashboard)
![Alt text](images/dashboard.png?raw=true "Analytics-Dashboards")

##### Analytics Monitoring for APIs

![Alt text](images/analytics-monitoring.png?raw=true "Monitoring Analytics")

##### Analytics Graphs for Events

![Alt text](images/analytics-graphs.png?raw=true "Graphs for Analytics")


#### Customize API Analytics 

By changing the following artifacts, you can point the API Operator to use the API Analytics which is deployed outside the Kuberentes cluster or anywhere which is accessible to the API Operator running Kubernetes cluster.

- Create two secrets for the analytics server as follows.

    1. Secret 1: Analytics certificate
    2. Secret 2: Include admin credentials (base64 encoded username and password) and secret name of the secret 1.
    
    Samples can be found in api-operator/apim-analytics-configs/apim_analytics_secret_template.yaml
    
- To enable analytics you can change the apim_analytics_conf.yaml analyticsEnabled to true. Give the name of the secret you created above in the analyticsSecret field value.

    Samples can be found api-operator/apim-analytics-configs/apim_analytics_conf.yaml

#### Clean Up

- Delete the created API (Find APIs `apictl get api --all-namespaces`)
    ```shell script
    >> apictl delete api <API_NAME> -n <NAMESPACE>
    ```
- Delete WSO2 APIM deployment and WSO2 Analytics deployment
    ```shell script
    >> apictl delete -Rf k8s-artifacts/wso2am-operator/api-portal-with-analytics
    ```
- Delete Analytics configs
    ```shell script
    >> apictl delete -f api-operator/apim-analytics-configs
    ```