## Developer Guide

### Build from source

The api-operator was built using the [operator-sdk (version v0.18.2)][operator_sdk].
Follow the [quick start guide][operator_sdk_quick_start] of operator-sdk for more information.

You should build the image when changes are added to the project. The steps to build the image are listed below. 

1.  After modifying the *_types.go files run the following command to update the generated code for that resource type
    ```sh
    >> operator-sdk generate k8s
    >> operator-sdk generate crds
    ```

1.  Build the api-operator image 
    ```sh
    >> operator-sdk build wso2/k8s-api-operator:v1.2.2
    ```

1.  Replace the image name in deploy/controller-artifacts/operator.yaml#L36

1.  Push it to a registry:
    ```sh
    >> docker push wso2/k8s-api-operator:v1.2.2
    ```

### Add new API and controller

1. Adding new custom resource definition
   ```sh
   >> operator-sdk add api --api-version=wso2.com/v1alpha1 --kind=<kind name>
   ```

1. Add a new Controller to the project
   ```sh
   >> operator-sdk add controller --api-version=wso2.com/v1alpha1 --kind=<kind name>
   ```

[operator_sdk]: https://github.com/operator-framework/operator-sdk/releases/tag/v0.18.2
[operator_sdk_quick_start]: https://v0-18-x.sdk.operatorframework.io/docs/golang/quickstart/