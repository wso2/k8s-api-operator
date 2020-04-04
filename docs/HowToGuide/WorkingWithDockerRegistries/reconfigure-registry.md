# Reconfigure Registry

You can reconfigure/change registry using the `apictl` tool. Use the `apictl change registry` command and follow the interactive session to reconfigure the registry.

```sh
>> apictl change registry
```

You can also use the non-interactive method to reconfigure the registry. Follow the same flags with the registry type [here](../install-api-operator-in-cicd.md) with the command `apictl change registry`.

Example usage:
```sh
Docker-Hub:
>> apictl change registry --registry-type=DOCKER_HUB --repository=docker.io/wso2 --username=john --password=*******