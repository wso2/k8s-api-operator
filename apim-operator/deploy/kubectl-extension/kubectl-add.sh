#Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
#
# WSO2 Inc. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

#!/bin/bash
# Please copy this file to /usr/local/bin and give executable permissions
# Sample usage : kubectl add api api-name --from-file=/home/harsha/Downloads/boomi/non-weborders/order_swagger.json --replicas=3
if [[ "$1" == "api" ]]
then

for var in "$@"
do
    if [[ $var == *"--from-file"* ]]; then
      fromFile=$var
    fi

    if [[ $var == *"--replicas"* ]]; then
      replicasArg=$var
    fi
done

IFS='=';

count=0
replicas=1

for i in $replicasArg;
do
    echo $i
    ((count++))
    if [[ $count == 2 ]]; then
          replicas=$i
    fi
done

apiName=$2

echo -e "\nDeleteting configmap if exists with name "$apiName
    kubectl delete configmap $2 -n=wso2-system

echo -e "\nCreating configmap with name "$apiName
    kubectl create configmap $2 $3 -n=wso2-system
    echo -e "\nGenerating a api kind"

cat << EOF > wso2_v1alpha1_api_cr.yaml
apiVersion: wso2.com/v1alpha1
kind: API
metadata:
 name: "${apiName}"
 namespace: wso2-system
spec:
 definition:
   configMapKeyRef:
     name: "${apiName}"
   replicas: ${replicas}
 mode: privateJet
EOF

    kubectl create -f wso2_v1alpha1_api_cr.yaml

else
    echo "Unknown command"
fi

echo "Completed"