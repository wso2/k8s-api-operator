# Working with Kaniko Additional Flags

[Kaniko](https://github.com/GoogleContainerTools/kaniko) is the tool which API operator uses to build the relevant API
image and push into the configured docker repository.

Mandatory flags for the Kaniko job are provided by the API Operator. You can refer the available additional flags for
the Kaniko [here](https://github.com/GoogleContainerTools/kaniko#additional-flags).

***Note*** : If you are HTTP registry, you do not need to add ***--insecure*** flag. It has also been handled by
API Operator.

If you wish to add these additional flags to the Kaniko job, please follow the instructions below.

1. Find the config map named "kaniko-arguments" (can be found inside
`<API-Operator-home>/api-operator/controller-artifacts/controller_conf.yaml`)
2. Include the [Kaniko additional flags](https://github.com/GoogleContainerTools/kaniko#additional-flags)
you need under ***KanikoArguments***. 

An example is shown below.
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kaniko-arguments
  namespace: wso2-system
data:
  kanikoArguments: |
    --no-push
    --skip-tls-verify
```

3. Apply the modified config map on the Kubernetes cluster.

From here onwards every Kaniko job will use the additional-flags you added above when building and pushing the image.