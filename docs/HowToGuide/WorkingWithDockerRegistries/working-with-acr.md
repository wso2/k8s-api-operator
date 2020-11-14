# Working with Azure Container Registry

You can use Azure Container Registry as a registry to push the built micro-gateway.

## Get Access Token

Follow https://docs.microsoft.com/en-us/azure/container-registry/container-registry-authentication#az-acr-login-with---expose-token
to get an Access Token for the ACR.

## Install API Operator

- Set the operator version as `v1.2.2` by executing following in a terminal.
    ```sh
    >> export WSO2_API_OPERATOR_VERSION=v1.2.2
    ```
- Execute the following command to install API Operator interactively and configure the repository to push the
  Microgateway image.
- Select "HTTPS" as the repository type.
- Use `<loginServer>` as the repository name.
- Enter `00000000-0000-0000-0000-000000000000` as the username.
- Enter `<accessToken>` as the password.
- Confirm the configuration are correct with entering "Y"

Sample Access Token.
```json
{
  "accessToken": "eyJhbGciOiJSUzI1NiIs[...]24V7wA",
  "loginServer": "myregistry.azurecr.io"
}
```

**Note:** It is also possible to use credentials of admin account of ACR.
For more info follow https://docs.microsoft.com/en-us/azure/container-registry/container-registry-authentication#admin-account.

```sh
>> apictl install api-operator
Choose registry type:
1: Docker Hub
2: Amazon ECR
3: GCR
4: HTTP Private Registry
5: HTTPS Private Registry
6: Quay.io
Choose a number: 1: 5
Enter repository name: myregistry.azurecr.io
Enter username: 00000000-0000-0000-0000-000000000000
Enter password: eyJhbGciOiJSUzI1NiIs[...]24V7wA

Repository: myregistry.azurecr.io
Username  : 00000000-0000-0000-0000-000000000000
Confirm configurations: Y:
```

```sh
Output:
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

## Try out
Try out [sample scenarios](../../GettingStarted/quick-start-guide.md#sample-scenarios) in the quick start guide.

## Clean up

- Delete images created by the operator in ACR repository.
- Uninstall the operator

```sh
>> apictl uninstall api-operator

Uninstall "api-operator" and all related resources: APIs, Securities, Rate Limitings and Target Endpoints
[WARNING] Remove the namespace: wso2-system
Are you sure: N: Y
```

```sh
Output:
Deleting kubernetes resources for API Operator
Removing namespace: wso2-system
This operation will take some minutes...
namespace "wso2-system" deleted
customresourcedefinition.apiextensions.k8s.io "apis.wso2.com" deleted
customresourcedefinition.apiextensions.k8s.io "securities.wso2.com" deleted
customresourcedefinition.apiextensions.k8s.io "ratelimitings.wso2.com" deleted
customresourcedefinition.apiextensions.k8s.io "targetendpoints.wso2.com" deleted
```