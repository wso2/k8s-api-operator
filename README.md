# k8s-apim-operator

How to try out the K8s API Manager Operator

Copy the the kubectl-add.sh to /usr/local/bin/

kubectl create -f deploy/namespace.yaml 

kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role_binding.yaml

Before deploying the role you have to make yourself as a cluster admin

kubectl create clusterrolebinding harsz89@gmail.com --clusterrole=cluster-admin --user=harsz89@gmail.com

kubectl create -f deploy/role.yaml

kubectl create -f deploy/crds/wso2_v1alpha1_targetendpoint_crd.yaml
kubectl create -f deploy/crds/wso2_v1alpha1_security_crd.yaml
kubectl create -f deploy/crds/wso2_v1alpha1_ratelimiting_crd.yaml
kubectl create -f deploy/crds/wso2_v1alpha1_api_crd.yaml

kubectl create -f deploy/operator.yaml

Modify the controller configurations and deploy
kubectl create -f deploy/controller_conf.yaml

Add your user name and password as a secret. A template is provided.
kubectl create -f deploy/docker_secret_template.yaml

Modify the below yaml and deploy, if you need to use analytics
kubectl create -f deploy/analytics_secret_template.yaml

kubectl create -f deploy/crds/wso2_v1alpha1_targetendpoint_cr.yaml

kubectl create -f deploy/crds/wso2_v1alpha1_security_cr.yaml

kubectl create -f deploy/crds/wso2_v1alpha1_ratelimiting_cr.yaml

kubectl create -f deploy/crds/wso2_v1alpha1_api_cr.yaml




Undeploy the changes

kubectl delete -f deploy/crds/wso2_v1alpha1_targetendpoint_cr.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_security_cr.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_ratelimiting_cr.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_api_cr.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/crds/app_v1alpha1_appservice_crd.yaml

kubectl delete -f deploy/controller_conf.yaml
kubectl delete -f deploy/docker_secret_template.yaml
kubectl delete -f deploy/analytics_secret_template.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_targetendpoint_crd.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_security_crd.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_ratelimiting_crd.yaml
kubectl delete -f deploy/crds/wso2_v1alpha1_api_crd.yaml
kubectl delete -f deploy/namespace.yaml 
