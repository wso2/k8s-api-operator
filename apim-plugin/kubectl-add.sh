#!/bin/bash
# Please copy this file to /usr/local/bin and give executable permissions
# Sample usage : kubectl add api api-name --from-file=/home/harsha/Downloads/boomi/non-weborders/order_swagger.json
if [[ "$1" == "api" ]]
then
    kubectl create configmap $2 $3
else
    echo "Unknown command"
fi

echo "Completed"