## Developer Guide

### Build from source

#### Prerequisites

- Maven
- Docker  
- [Operator-sdk v0.18.2]

    Follow the [quick start guide][operator_sdk_quick_start] of operator-sdk for more information.

#### Build the source and the Docker image

```sh
>> mvn clean install
```

- The docker image of the operator gets created. 

#### Push the docker image (Optional)

- Create a settings.xml file in ~/.m2 directory of Maven.
- Include the following in the settings.xml and replace USERNAME AND PASSWORD fields

```code
<settings xmlns="http://maven.apache.org/SETTINGS/1.1.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.1.0 http://maven.apache.org/xsd/settings-1.1.0.xsd">
  <servers>
  	<server>
   		<id>docker.io</id>
   		<username>USERNAME</username>
   		<password>PASSWORD</password>
  	</server>
  </servers>
</settings>
```

- Execute the following command to push the docker image to the repository
```sh
>> mvn dockerfile:push
```

### Advanced Configurations

1.  After modifying the *_types.go files run the following command to update the generated code for that resource type
    ```sh
    >> operator-sdk generate k8s
    >> operator-sdk generate crds
    ```

1.  Build the api-operator image (Optional: Use mvn clean install)
    ```sh
    >> operator-sdk build wso2/k8s-api-operator:2.0.1
    ```

1.  Replace the image name in deploy/controller-artifacts/operator.yaml#L36

1.  Push it to a registry:
    ```sh
    >> docker push wso2/k8s-api-operator:2.0.1
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

[Operator-sdk v0.18.2]: https://github.com/operator-framework/operator-sdk/releases/tag/v0.18.2
[operator_sdk_quick_start]: https://v0-18-x.sdk.operatorframework.io/docs/golang/quickstart/

### Running Unit Tests

- Execute the following command to run unit tests

   ```sh
   >> gotest -v -covermode=count -coverprofile=coverage.out ./pkg/...
   ```
  
- Check coverage locally

   ```sh
   >> go tool cover -html=coverage.out
   ```

### How to Debug API Operator

Copy following certs in the directory [build/security](build/security) to `/home/wso2/security/` (or update the cert directory in
[this file](pkg/envoy/server/api/restserver/configure_restapi.go) to the path [build/security](build/security))

Install API Operator from distribution and delete the API Operator deployment.

```sh
kubectl delete deploy api-operator
```

Execute `cmd/manager/main.go` file in debug mode.
