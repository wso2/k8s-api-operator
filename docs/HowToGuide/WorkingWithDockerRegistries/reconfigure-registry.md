## Reconfigure Registry

You can reconfigure/change registry using the `apictl` tool. Use the `apictl change registry`
command and follow the interactive session to reconfigure the registry.

```sh
>> apictl change registry
Choose registry type:
1: Docker Hub
2: Amazon ECR
3: GCR
4: HTTP Private Registry
5: HTTPS Private Registry
6: Quay.io
Choose a number: 1: 1
Enter repository name: docker.io/john
Enter username: john
Enter password:
```

You can also use the non-interactive method to reconfigure the registry.
Use the same flags, registry types used in ["Install API Operator in CI/CD"](../install-api-operator-in-cicd.md) to 
reconfigure the registry.

Example usage:
```sh
Docker-Hub:
>> apictl change registry \
            --registry-type=DOCKER_HUB \
            --repository=docker.io/<REPO_NAME> \
            --username=<USER_NAME> \
            --password=<PASSWORD>

HTTPS Private Registry:
>> apictl change registry \
            --registry-type=HTTPS \
            --repository=<REPOSITORY> \
            --username=<USER_NAME> \
            --password=<PASSWORD>
```