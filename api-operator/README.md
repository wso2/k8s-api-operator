## Developer Guide

### Build from source

The api-operator was built using the operator-sdk CLI (version v0.17.1) tool. You can follow the [user guide][operator_sdk_user_guide] of operator-sdk CLI tool for 
more information. 

You should build the image when changes are added to the project. The steps to build the image are listed below. 

1.  After modifying the *_types.go files run the following command to update the generated code for that resource type
    ```sh
    >> operator-sdk generate k8s
    >> operator-sdk generate crds
    ```

1.  Build the api-operator image 
    ```sh
    >> operator-sdk build wso2/k8s-api-operator:v1.2.0-alpha
    ```

1.  Replace the image name in deploy/controller-artifacts/operator.yaml:

1.  Push it to a registry:
    ```sh
    >> docker push wso2/api-operator:v1.2.0-alpha
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

[operator_sdk_user_guide]: https://github.com/operator-framework/operator-sdk/blob/v0.17.1/doc/user-guide.md