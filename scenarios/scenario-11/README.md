## Scenario 11 - Enabling Analytics and Monitor Dashboard in APIM Operator
- This scenario describes how to enable analytics in the apim-operator and monitor analytics in a dashboard

This setup provides resources to deploy WSO2 API Manager 3.0.0 and WSO2 APIM Analytics in kubernetes cluster and configure them with WSO2 Microgateway using k8s CRD operator.
 
Few databases of WSO2 API Manager and Analytics need to be shared with each other and hence we have provided resources to deploy mysql in kubernetes and configure the necessary databases in mysql.

- Create a secret with your base64 encoded username, password and the secret with the analytics certificate
- To enable analytics you can change the apim_analytics_conf.yaml analyticsEnabled to true. Give the name of the secret you created in step 1 in the analyticsSecret field value

Note: A secret with the default certificate is provided here

Scenario10

Navigate to <k8s-CRD-HOME>/scenarios/scenario-11 directory.
Execute the below commands to deploy database artifacts, api and analytics artifacts.

```
apimctl apply -f wso2-namespace.yaml
apimctl apply -f ./apim-analytics-configmaps
apimctl apply -f ./apim-analytics-controller-conf
apimctl apply -f ./mysql-artifacts
apimctl apply -f ./apim-analytics-setup
```

Then follow scenario-2 instructions to deploy a sample api with a microgateway
