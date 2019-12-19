## OperatorHub Resources

[OperatorHub.io](https://operatorhub.io) is a place where K8s community share Operators. API Operator is available in OperatorHub and is compatible with Operator Lifecycle Manager.

For more information please visit [OperatorHub](https://operatorhub.io/operator/apim-operator) 

The Operator Lifecycle Manager (OLM) helps users install, update, and manage the lifecycle of all Operators and their associated services running across their clusters. It is part of the Operator Framework, an open source toolkit designed to manage Kubernetes native applications (Operators) in an effective, automated, and scalable way.

### Important concepts related to OLM

#### CSV
A ClusterServiceVersion (CSV) is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator in a cluster. It contains the metadata such as name, version, icon, required resources, installation, etc.
#### Subscription
A subscription keeps CSVs up to date by tracking a channel in a package.
#### OperatorGroup
An OperatorGroup selects a set of target namespaces in which to generate required RBAC access for its member operators
#### CatalogSource
A CatalogSource is a repository of CSVs, CRDs, and packages that define an application.

For an operator to be managed successfully using OLM, the above 4 resources should be present in the cluster.

### Documentation

1. [Default install](default-install/README.md)
2. [Install in a specific namespace](namespace-install/README.md)
3. [Creating new version in operatorhub](creating-new-version/README.md)
4. [Local testing](local-testing/README.md)
