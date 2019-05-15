#!/bin/bash
# Please copy this file to /usr/local/bin and give executable permissions
# Sample usage : kubectl add api api-name --from-file=/home/harsha/Downloads/boomi/non-weborders/order_swagger.json
if [[ "$1" == "api" ]]
then

apiName=$2
echo -e "\nDeleteting configmap if exists with name "$apiName
    kubectl delete configmap $2

echo -e "\nCreating configmap with name "$apiName
    kubectl create configmap $2 $3
    echo -e "\nGenerating a api kind"

cat << EOF > wso2_v1alpha1_api_cr.yaml
apiVersion: wso2.com/v1alpha1
kind: API
metadata:
 name: "${apiName}"
spec:
 definition:
   configMapKeyRef:
     name: "${apiName}"
 mode: shared
EOF

    kubectl create -f wso2_v1alpha1_api_cr.yaml

else
    echo "Unknown command"
fi

echo "Completed"