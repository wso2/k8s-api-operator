# Working with Google Container Registry

You can use Google Container Registry as a registry to push the built micro-gateway.

## Download Service Account Key JSON File

Follow the [cloud.google.com documentation](https://cloud.google.com/docs/authentication/getting-started#creating_a_service_account) to create a service account key JSON File and download it to your local file system.

## Install API Operator

- Execute the following command to install API Operator interactively and configure repository to push the microgateway image.
- Select "GCR" as the repository type.
- Enter the file path of the downloaded service account key JSON File.
- Confirm configuration are correct with entering "Y"

```sh
>> apictl install api-operator
Choose registry type:
1: Docker Hub (Or others, quay.io, HTTPS registry)
2: Amazon ECR
3: GCR
4: HTTP Private Registry
Choose a number: 1: 3
GCR service account key json file: /path/to/gcr/service/account/key/file.json

GCR service account key json file: /path/to/gcr/service/account/key/file.json
Confirm configurations: Y:
```

Output:
```sh
[Installing OLM]
customresourcedefinition.apiextensions.k8s.io/clusterserviceversions.operators.coreos.com created
...

[Installing API Operator]
subscription.operators.coreos.com/my-api-operator created
[Setting configs]
namespace/wso2-system created
...

[Setting to K8s Mode]
```