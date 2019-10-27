## Scenario 11 - Enabling Analytics for managed API


- This scenario describes how to enable analytics in the apim-operator and monitor analytics in the analytics dashboard

This setup provides resources to deploy WSO2 API Manager 3.0.0 and WSO2 APIM Analytics 3.0.0 in the Kubernetes cluster and configure them with WSO2 Microgateway using k8s CRD operator.
 
Few databases of WSO2 API Manager and Analytics need to be shared with each other and hence we have provided resources to deploy mysql in kubernetes and configure the necessary databases in mysql.


To try out the scenario navigate to wso2am-k8s-crds-1.0.0 directory.

###### Step 1: Deploy and configure API Portal and Analytics Dashbaord

- Execute the following commands based on your Kubernetes environment

<details><summary>If you are using GKE, Docker for desktop, etc.(Load Balancer Service Type)</summary>

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
</details>



<details><summary>If you are using minikube (NodePort Service Type)</summary>

```
>> apictl apply -f k8s-artifacts/api-portal-with-analytics-for-minikube/wso2-namespace.yaml

---
namespace/wso2 created  
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics-for-minikube/mysql/

---
configmap/mysql-dbscripts created
deployment.apps/wso2apim-with-analytics-mysql-deployment created
service/wso2apim-with-analytics-rdbms-service created
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics-for-minikube/api-analytics/

---
configmap/dash-bin created
configmap/dash-conf created
deployment.apps/wso2apim-dashboard-analytics-deployment created
service/wso2apim-dashboard-analytics-service created
---

>> apictl apply -f k8s-artifacts/api-portal-with-analytics-for-minikube/api-portal/

---
configmap/apim-conf created
deployment.apps/wso2apim created
service/wso2apim created
---
```
</details>

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
NAME                                    TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)                                                                            AGE
wso2apim                                LoadBalancer   10.0.29.36    104.154.100.37   8280:30858/TCP,8243:30536/TCP,9763:31828/TCP,9443:30248/TCP                        3m20s
wso2apim-analytics-service              ClusterIP      10.0.18.193   <none>           7612/TCP,7712/TCP,9444/TCP,9091/TCP,7071/TCP,7444/TCP,7575/TCP,7576/TCP,7577/TCP   3m24s
wso2apim-dashboard-analytics-service    LoadBalancer   10.0.28.250   35.193.108.59    9643:30620/TCP                                                                     3m28s
wso2apim-with-analytics-rdbms-service   ClusterIP      10.0.17.109   <none>           3306/TCP
---
```

**Note:** To access the API portal and Analytics dashboard, add a host mapping entry to the /etc/hosts file. Use the above external IP of each service as the IP address to the host wso2apim and wso2apim-analytics.

```
<EXTERNAL-IP_of_service_wso2apim>  wso2apim
<EXTERNAL-IP_of_service_wso2apim-dashboard-analytics-service>  wso2apim-analytics
```

- For Docker for Mac “external-IP” should be “localhost” and port is 9443 <br>

    **API Portal** - https://wso2apim:9443/devportal <br>
    **API Analytics Dashbaord** - https://wso2apim-analytics:9643/analytics-dashboard

- For Minikube “external-ip” is Minikube VM IP (Use minikube ip command to get the IP) and port is 32001 <br>
	 
    **API Portal** - https://wso2apim:32001/devportal <br>
    **API Analytics Dashbaord** - https://wso2apim-analytics:32201/analytics-dashboard


###### Step 2: Enable API Analytics in the API Operator

- If you haven't deployed the API Operator please follow the quick start guide in root readme.
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


#### Customize API Analytics 

By changing the following artifacts, you can point the API Operator to use the API Analytics which is deployed outside the Kuberentes cluster or anywhere which is accessible to the API Operator running Kubernetes cluster.

- Create two secrets for the analytics server as follows.

    1. Secret 1: Analytics certificate
    2. Secret 2: Include admin credentials (base64 encoded username and password) and secret name of the secret 1.
    
    Samples can be found in apim-operator/apim-analytics-configs/apim_analytics_secret_template.yaml
    
- To enable analytics you can change the apim_analytics_conf.yaml analyticsEnabled to true. Give the name of the secret you created above in the analyticsSecret field value.

    Samples can be found apim-operator/apim-analytics-configs/apim_analytics_conf.yaml

