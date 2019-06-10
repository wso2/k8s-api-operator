# k8s-apim-operator

##### Navigate to the k8s-apim-operator/apim-operators/ directory and execute the following command

###### Deploy k8s client extension
- Give executable permission to the extension file <br /> 
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-add.sh***
- Copy the ***/deploy/kubectl-extension/kubectl-add.sh*** to ***/usr/local/bin/***
    - ___cp ./deploy/kubectl-extension/kubectl-add.sh /usr/local/bin___

###### Deploy k8s CRD artifacts

> ##### Before deploying the role you have to make yourself as a cluster admin. (Replace "email-address" with the proper value)
- *kubectl create clusterrolebinding email-address --clusterrole=cluster-admin --user=email-address*

> ##### Deploying namespace, roles/role binding and service account associated with operator
- _for i in ./deploy/controller-artifacts/*yaml; do kubectl apply -f $i; done_

> ##### Deploying CRD for API, Target endpoint, Security, Ratelimiting
- _for i in ./deploy/crd/*yaml; do kubectl apply -f $i; done_


> ##### Deploying controller level configuration
>> "controller-configs" contains the configuration user would have to
change.  
- _for i in ./deploy/controller-configs/*yaml; do kubectl apply -f $i; done_

> ###### Deploy sample custom resources on the kubernetes cluster
- _for i in ./deploy/sample-crs/*yaml; do kubectl apply -f $i; done_

> ##### Deploying an API in K8s cluter

- Download sample API definition (swagger files) from [product micro-gateway sample](https://github.com/wso2/product-microgateway/tree/master/samples) github location.
- Execute the following command to deploy the API
- *kubectl add api "api_name" --from-file="location to the api swagger definition"*
- ex: ***kubectl add api petstorebasic --from-file=petstore_basic.yaml***
> ###### Undeploy the changes (one by one)

- *kubectl delete -f deploy/crds/wso2_v1alpha1_targetendpoint_cr.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_security_cr.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_ratelimiting_cr.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_api_cr.yaml*
- *kubectl delete -f deploy/operator.yaml*
- *kubectl delete -f deploy/role.yaml*
- *kubectl delete -f deploy/role_binding.yaml*
- *kubectl delete -f deploy/service_account.yaml*
- *kubectl delete -f deploy/crds/app_v1alpha1_appservice_crd.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_targetendpoint_crd.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_security_crd.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_ratelimiting_crd.yaml*
- *kubectl delete -f deploy/crds/wso2_v1alpha1_api_crd.yaml*
- *kubectl delete -f deploy/namespace.yaml* 
- *kubectl delete -f deploy/controller_conf.yaml*
- *kubectl delete -f deploy/docker_secret_template.yaml*
- *kubectl delete -f deploy/analytics_secret_template.yaml*
