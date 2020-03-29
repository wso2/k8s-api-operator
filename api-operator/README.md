#### Developer Guide

**Building the image**

The api-operator was built using the operator-sdk CLI tool. you can follow the [user guide][operator_sdk_user_guide] of operator-sdk CLI tool for 
more information. 

you should build the image when changes are added to the project. The steps to build the image are listed below. 

1.  After modifying the *_types.go files run the following command to update the generated code for that resource type
    ```
    operator-sdk generate k8s
    ```
2.  Build the api-operator image 
    ```
    operator-sdk build wso2/k8s-api-operator:v1.1.0
    ```
3.  Replace the image name in deploy/controller-artifacts/operator.yaml:

4.  Push it to a registry:
    ```
    docker push wso2/api-operator:v0.6
    ```
    
**Additional Commands**

1. Adding new custom resource definition
    ```
    operator-sdk add api --api-version=wso2.com/v1alpha1 --kind=<kind name>
    ```
2. Add a new Controller to the project
   ```
   operator-sdk add controller --api-version=wso2.com/v1alpha1 --kind=<kind name>
   ```
    
[operator_sdk_user_guide]:https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md