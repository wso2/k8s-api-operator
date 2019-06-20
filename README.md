# k8s-apim-operator

##### Navigate to the k8s-apim-operator/apim-operators/ directory and execute the following command

##### Deploy k8s client extensions
- Give executable permission to the extension file <br /> 
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-add.sh***
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-update.sh***
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-remove.sh***
    -  ***chmod +x ./deploy/kubectl-extension/kubectl-show.sh***
- Copy the extensions to ***/usr/local/bin/***
    - ___cp ./deploy/kubectl-extension/kubectl-add.sh /usr/local/bin___
    - ___cp ./deploy/kubectl-extension/kubectl-update.sh /usr/local/bin___
    - ___cp ./deploy/kubectl-extension/kubectl-remove.sh /usr/local/bin___
    - ___cp ./deploy/kubectl-extension/kubectl-show.sh /usr/local/bin___

###### Note:
- User may have to remove '.sh' from the extension depending on the kubernetes environment. Use the following additional commands for that.
    - mv /usr/local/bin/kubectl-add.sh /usr/local/bin/kubectl-add
    - mv /usr/local/bin/kubectl-update.sh /usr/local/bin/kubectl-update
    - mv /usr/local/bin/kubectl-remove.sh /usr/local/bin/kubectl-remove
    - mv /usr/local/bin/kubectl-show.sh /usr/local/bin/kubectl-show
##### Deploy k8s CRD artifacts

> ##### Before deploying the role you have to make yourself as a cluster admin. (Replace "email-address" with the proper value)
- *kubectl create clusterrolebinding email-address --clusterrole=cluster-admin --user=email-address*

> ##### Deploying namespace, roles/role binding and service account associated with operator
- _for i in ./deploy/controller-artifacts/*yaml; do kubectl apply -f $i; done_

> ##### Deploying CRD for API, Target endpoint, Security, Ratelimiting
- _for i in ./deploy/crd/*yaml; do kubectl apply -f $i; done_


> ##### Deploying controller level configuration
>> "controller-configs" contains the configuration user would have to change. Modify the controller_conf.yaml with the needed values and enter the ***user's docker registry***. Enter the base 64 encoded username and password of the user's docker registry into the docker_secret_template.yaml.
- _for i in ./deploy/controller-configs/*yaml; do kubectl apply -f $i; done_

> ##### Deploy sample custom resources on the kubernetes cluster (Optional)
- _for i in ./deploy/sample-crs/*yaml; do kubectl apply -f $i; done_

> ##### Deploying an API in K8s cluter

- Download sample API definition (swagger files) from [product micro-gateway sample](https://github.com/wso2/product-microgateway/tree/master/samples) github location.
- Execute the following command to deploy the API
    - *kubectl add api "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" -n="desired namespace"*
    - ex: ***kubectl add api petstorebasic --from-file=petstore_basic.yaml --replicas=2 -n=wso2***

- Execute the following command to update the API
    - *kubectl update api "api_name" --from-file="location to the api swagger definition" --replicas="number of replicas" -n="desired namespace"*
    - ex: ***kubectl update api petstorebasic --from-file=petstore_basic.yaml --replicas=1 -n=wso2***

- Execute the following command to view the details of the API
    - *kubectl show api "api_name" -n="desired namespace"*
    - ex: ***kubectl show api petstorebasic -n=wso2***

- Execute the following command to remove the API
    - *kubectl remove api "api_name" -n="desired namespace"*
    - ex: ***kubectl remove api petstorebasic -n=wso2***

- Execute the following command to remove the k8s operator
    - *kubectl remove operator


- Note:
- If the namespace is not provided, default namespace will be used. The namespace used in all the commands must match.


##### Incorporating analytics to the k8s operator

- To enable analytics, modify the analytics-config configmap given in the ./deploy/apim-analytics-configs/apim-analytics-conf.yaml and set the field analyticsEnabled to "true". The other parameters also can be modified with required values.
- Create a secret with the public certificate of the wso2am-analytics server and provide the name of the created secret along with the username and password to the wso2am-analytics server (all fields must be base 64 encoded). Use the template provided for analytics-secret in apim_analytics_secret_template.yaml

##### Using OAuth2 
- To enable OAuth2, deploy a security kind with the desired parameters. A sample is provided in ./deploy/sample-crs/wso2_v1alpha1_security_cr.yaml. The certificate of the wso2am server and credentials should be included in secrets and the secret names should be entered to the security kind.

- Note:
- Modify the configurations related to wso2am using the template provided in ./deploy/apim-analytics-configs/apim-analytics-conf.yaml : apim-config configmap.

