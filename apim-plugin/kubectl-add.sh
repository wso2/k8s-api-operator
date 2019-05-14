#!/bin/bash
# Please copy this file to /usr/local/bin and give executable permissions
# Sample usage : kubectl add api api-name --from-file=/home/harsha/Downloads/boomi/non-weborders/order_swagger.json
if [[ "$1" == "api" ]]
then
    kubectl create configmap $2 $3

    echo -e "\nGenerating a api kind"

      # Generate the file
      cat <<EOF | kubectl apply -f -
        apiVersion: wso2.com/v1alpha1
        kind: API
        metadata:
        name: $2
        spec:
          definition:
            configMapKeyRef:
                name: $2
          mode: shared
      EOL

else
    echo "Unknown command"
fi

echo "Completed"