## Creating a new version

#### Prerequisites

[quay.io](https://quay.io/) account

[operator-sdk v0.18.2](https://github.com/operator-framework/operator-sdk/releases/tag/v0.18.2) 

#### Steps

1. Create an operator bundle compatible with operator lifecycle manager. 
- [Bundle editor](https://operatorhub.io/bundle) can be used for creation of the operator bundle.
- Add the CRDs, role, role binding, service account and deployment of the api operator
- Include the description of the operator in metadata
- Include a quadratic image of the logo
- This will create a cluster service version (csv is inside the operator bundle)
- Download the operator bundle

This can also be done using the operator-sdk commands using command line. Navigate into the `api-operator` directory and execute the following command.

```shell script
>> operator-sdk generate bundle --version 1.2.2
```
This is the basic command to create a bundle.
For more information regarding additional commands and flags visit [here](https://sdk.operatorframework.io/docs/olm-integration/legacy/generating-a-csv/). 

2. You can preview the csv file created in the 1st step using [this](https://operatorhub.io/preview)

3. Add the directory with the new version to the api-operator bundle available [here](/../api-operator) and the bundle should be in the below format.

```
   api-operator
   ├── 1.0.1
   │   ├── api-operator.v1.0.1.clusterserviceversion.yaml
   │   ├── apis.wso2.com.crd.yaml
   │   ├── ratelimitings.wso2.com.crd.yaml
   │   ├── securities.wso2.com.crd.yaml
   │   └── targetendpoints.wso2.com.crd.yaml
   ├── 1.0.2
   │   ├── api-operator.v1.0.2.clusterserviceversion.yaml
   │   ├── apis.wso2.com.crd.yaml
   │   ├── ratelimitings.wso2.com.crd.yaml
   │   ├── securities.wso2.com.crd.yaml
   │   └── targetendpoints.wso2.com.crd.yaml
   └── api-operator.package.yaml
```
Note:
- api-operator.package.yaml contains the channels and the latest version available in those channel.
If you want to make few versions available (Ex: 1.0.1 and 1.0.2 to be available to download), then you can put the new version under a new channel like below.
```yaml
packageName: api-operator
channels:
  - name: stable
    currentCSV: api-operator.v1.0.1
  - name: alpha
    currentCSV: api-operator.v1.0.2
defaultChannel: stable
```

- If you want the subscribed customers to have a rolling update of the new version, put the new version in the same channel and put the old version in the 'replaces' field of the csv.
- If you are creating a new channel remove the 'replaces' field in the csv
 
4. Install OLM
```shell script
>> kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/crds.yaml
>> kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/olm.yaml
```

5. Deploy operator-marketplace. (operator-marketplace operator is needed only for local testing)
```shell script
>> git clone https://github.com/operator-framework/operator-marketplace.git
>> kubectl apply -f operator-marketplace/deploy/upstream/
```
5. Install operator courier
```shell script
>> pip3 install operator-courier
```

6. Verify your bundle
```shell script
>> operator-courier --verbose verify --ui_validate_io api-operator

INFO:operatorcourier.verified_manifest:The source directory is in nested structure.
INFO:operatorcourier.verified_manifest:Parsing version: 1.1.0
INFO: Validating bundle. []
INFO: Validating custom resource definitions. []
INFO: Evaluating crd apis.wso2.com [1.1.0/apis.wso2.com.crd.yaml]
INFO: Evaluating crd securities.wso2.com [1.1.0/securities.wso2.com.crd.yaml]
INFO: Evaluating crd targetendpoints.wso2.com [1.1.0/targetendpoints.wso2.com.crd.yaml]
INFO: Evaluating crd ratelimitings.wso2.com [1.1.0/ratelimitings.wso2.com.crd.yaml]
INFO: Validating cluster service versions. [1.1.0/ratelimitings.wso2.com.crd.yaml]
INFO: Evaluating csv api-operator.v1.1.0 [1.1.0/api-operator.v1.1.0.clusterserviceversion.yaml]
INFO: Validating packages. [1.1.0/api-operator.v1.1.0.clusterserviceversion.yaml]
INFO: Evaluating package api-operator [api-operator/api-operator.package.yaml]
INFO: Validating cluster service versions for operatorhub.io UI. [api-operator/api-operator.package.yaml]
INFO: Evaluating csv api-operator.v1.1.0 [api-operator/api-operator.package.yaml]

```

7. Login to quay.io account

```shell script
$ ./operator-courier/scripts/get-quay-token
Username: johndoe
Password:
{"token": "basic abcdefghijkl=="}

export QUAY_TOKEN="basic abcdefghijkl=="
```
8. Push the bundle to quay.io

```shell script
export OPERATOR_DIR=api-operator/
export QUAY_NAMESPACE=johndoe
export PACKAGE_NAME=api-operator
export PACKAGE_VERSION=1.0.1
export TOKEN=$QUAY_TOKEN
>> operator-courier --verbose push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
```
9. Go to your quay.io account and check if the bundle is available in the **applications** tab.

**Note:**
When you push the bundle, it will be private in your quay.io. Please make sure to change the repository to **public**

10. Test the operator locally following the instructions [here](../local-testing/README.md)

11. Contribute to [operatorhub](https://github.com/operator-framework/community-operators)

- For Kubernetes operators which are available in operatorhub.io, contribute to [upstream operators](https://github.com/operator-framework/community-operators/tree/master/upstream-community-operators)
- For Openshift community operators, contribute to [community operators](https://github.com/operator-framework/community-operators/tree/master/community-operators)

