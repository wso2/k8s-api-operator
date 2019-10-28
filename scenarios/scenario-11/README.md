## Scenario 11 - Enabling Analytics for managed API


- This scenario describes how to enable analytics in the apim-operator and monitor analytics in the analytics dashboard

This setup provides resources to deploy WSO2 API Manager 3.0.0 and WSO2 APIM Analytics 3.0.0 in the Kubernetes cluster and configure them with WSO2 Microgateway using k8s CRD operator.
 
Few databases of WSO2 API Manager and Analytics need to be shared with each other and hence we have provided resources to deploy mysql in kubernetes and configure the necessary databases in mysql.


To try out the scenario navigate to wso2am-k8s-crds-1.0.0 directory.

###### Step 1: Deploy and configure API Portal and Analytics Dashbaord

**Note:** If you have already deployed the API portal without enabling analytics, please remove it by executing the below command.

```
>> kubectl delete k8s-artifacts/api-portal/
---
namespace "wso2" deleted
configmap "apim-conf" deleted
deployment.apps "wso2apim" deleted
service "wso2apim" deleted
```

- Execute the following commands to deploy the API Portal and Analytics Dashboard

```
>> apictl apply -f k8s-artifacts/api-portal-with-analytics/wso2-namespace.yaml

---
namespace/wso2 created  
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics/mysql/

---
configmap/mysql-dbscripts created
deployment.apps/wso2apim-with-analytics-mysql-deployment created
service/wso2apim-with-analytics-rdbms-service created
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics/api-analytics/

---
configmap/dash-bin created
configmap/dash-conf created
deployment.apps/wso2apim-dashboard-analytics-deployment created
service/wso2apim-dashboard-analytics-service created
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics/api-portal/

---
configmap/apim-conf created
deployment.apps/wso2apim created
service/wso2apim created
---
```

You can verify the installation by checking the pods and services as follows.

```
>> apictl get pods -n wso2

---
NAME                                                        READY   STATUS    RESTARTS   AGE
wso2apim-76b4cd8974-gpdz7                                   1/1     Running   0          3m12s
wso2apim-analytics-deployment-cdc8db56b-zv6qp               1/1     Running   0          3m17s
wso2apim-dashboard-analytics-deployment-79fb44f4b8-p49km    1/1     Running   0          3m20s
wso2apim-with-analytics-mysql-deployment-749dd5fb7b-fh7cd   1/1     Running   0          3m28s
---

>> apictl get services -n wso2

---

NAME                                    TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                                                                            AGE
wso2apim                                NodePort    10.0.28.252   <none>        30838:32004/TCP,30801:32003/TCP,32321:32002/TCP,32001:32001/TCP                    69m
wso2apim-analytics-service              ClusterIP   10.0.18.144   <none>        7612/TCP,7712/TCP,9444/TCP,9091/TCP,7071/TCP,7444/TCP,7575/TCP,7576/TCP,7577/TCP   69m
wso2apim-dashboard-analytics-service    NodePort    10.0.24.27    <none>        32201:32201/TCP                                                                    70m
wso2apim-with-analytics-rdbms-service   ClusterIP   10.0.23.125   <none>        3306/TCP                                                                           70m
---
```

**Note:** To access the API portal and Analytics dashboard, add host mapping entries to the /etc/hosts file. As we have exposed the services in Node Port type, you can use the IP address of any Kubernetes node.


```
<Any K8s Node IP>  wso2apim
<Any K8s Node IP>  wso2apim-analytics
```

- For Docker for Mac use "localhost" for the K8s node IP 
- For Minikube, use minikube ip command to get the K8s node IP
	 
    **API Portal** - https://wso2apim:32001/devportal <br>
    **API Analytics Dashbaord** - https://wso2apim-analytics:32201/analytics-dashboard


###### Step 2: Enable API Analytics in the API Operator

- If you haven't deployed the API Operator please follow the quick start guide in root readme and follow step 1,2 and 4.
- By deploying the analytics configmaps, you can enable analytics as follows.

```
>> apictl apply -f apim-operator/apim-analytics-configs/

---
configmap/analytics-config created
secret/analytics-secret created
secret/wso2analytics300-secret created
---
```

###### Step 3: Deploy an API

Please follow the scenario 1 readme and deploy the API.

Use the API Analytics dashboard url which configured in Step 1 to browse the analytics information for the API.

You will be able to monitor the analytics as shown in below images.

![Alt text](images/Analytics-Dashboard.png?raw=true "Title")

![Alt text](images/Developer-Analytics.png?raw=true "Title")

![Alt text](images/Publisher-Analytics.png?raw=true "Title")


#### Customize API Analytics 

By changing the following artifacts, you can point the API Operator to use the API Analytics which is deployed outside the Kuberentes cluster or anywhere which is accessible to the API Operator running Kubernetes cluster.

- Create two secrets for the analytics server as follows.

    1. Secret 1: Analytics certificate
    2. Secret 2: Include admin credentials (base64 encoded username and password) and secret name of the secret 1.
    
    Samples can be found in apim-operator/apim-analytics-configs/apim_analytics_secret_template.yaml
    
- To enable analytics you can change the apim_analytics_conf.yaml analyticsEnabled to true. Give the name of the secret you created above in the analyticsSecret field value.

    Samples can be found apim-operator/apim-analytics-configs/apim_analytics_conf.yaml

