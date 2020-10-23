# Install API Operator in CI/CD

API Controller, **apictl** tool enable you to install api-operator and configure registry in interactive mode and
non-interactive mode as well. For CI/CD automation you can use the non-interactive feature based on the registry type.

```sh
Flags:
  -f, --from-file string       Path to API Operator config directory or file
  -c, --key-file string        Credentials file
  -p, --password string        Password of the given user
      --password-stdin         Prompt for password of the given user in the stdin
  -R, --registry-type string   Registry type: DOCKER_HUB | AMAZON_ECR | GCR | HTTP
  -r, --repository string      Repository name or URI
  -u, --username string        Username of the repository
```

For offline installation use the `--from-file` flag to point the operator configs downloaded from
the [release](https://github.com/wso2/k8s-api-operator/releases). The configs pointing with this flag can be
- a file with all configs
- a directory with all configs files
- an URI

The flag `--from-file` can be used with any registry type.

## 1. Install with Configuring a Registry

If it is needed to reconfigure registry after installing API Operator, you can follow the document
["Reconfigure Registry"](WorkingWithDockerRegistries/reconfigure-registry.md).

- Set the operator version as `v1.2.2` by executing following in a terminal.
    ```sh
    >> export WSO2_API_OPERATOR_VERSION=v1.2.2
    ```

### 1.1. Docker Hub

Registry type: **DOCKER_HUB**

```sh
Format:
>> apictl install api-operator \
            --registry-type=DOCKER_HUB \
            --repository=<REPOSITORY> \
            --username=<USER_NAME> \
            --password=<PASSWORD>
```
***Note:*** The flag `--password-stdin` can be used instead of `--password` if you want to read password from
standard input

```sh
Docker-Hub:
>> apictl install api-operator \
            --registry-type=DOCKER_HUB \
            --repository=docker.io/wso2 \
            --username=john \
            --password=*******
```

### 1.2. Amazon ECR

Registry type: **AMAZON_ECR**

```sh
Format:
>> apictl install api-operator \
            --registry-type=AMAZON_ECR \
            --repository=<REPOSITORY_URI> \
            --key-file=<AWS_CREDENTIALS_FILE>

Example:
>> apictl install api-operator \
            --registry-type=AMAZON_ECR \
            --repository=111222333444.dkr.ecr.us-east-1.amazonaws.com/my-ecr-repo \
            --key-file=/Users/wso2/.aws/credentials
```

### 1.3. GCR

Registry type: **GCR**

```sh
Format:
>> apictl install api-operator \
            --registry-type=GCR \
            --key-file=<PATH_TO_GCR_SERVICE_ACCOUNT_KEY_FILE_JSON>

Example:
>> apictl install api-operator \
            --registry-type=GCR \
            --key-file=/path/to/gcr/service/account/key/file.json
```

### 1.4. Azure Container Registry

Registry type: **HTTPS**

Get an Access Token for the Azure ACR by following https://docs.microsoft.com/en-us/azure/container-registry/container-registry-authentication#az-acr-login-with---expose-token.

- Use `<loginServer>` as the repository name
- Username:  `00000000-0000-0000-0000-000000000000`
- Password: `<accessToken>`

```sh
Format:
>> apictl install api-operator \
            --registry-type=HTTPS \
            --repository=<LOGIN_SERVER> \
            --username=00000000-0000-0000-0000-000000000000 \
            --password=<ACCESS_TOKEN>
```

Sample Access Token
```json
{
  "accessToken": "eyJhbGciOiJSUzI1NiIs[...]24V7wA",
  "loginServer": "myregistry.azurecr.io"
}
```

Example:
```sh
>> apictl install api-operator \
            --registry-type=HTTPS \
            --repository=myregistry.azurecr.io \
            --username=00000000-0000-0000-0000-000000000000 \
            --password=eyJhbGciOiJSUzI1NiIs[...]24V7wA
```

**Note:** It is also possible to use credentials of admin account of ACR.
For more info follow https://docs.microsoft.com/en-us/azure/container-registry/container-registry-authentication#admin-account.

### 1.5. HTTP Private Registry

Registry type: **HTTP**

```sh
Format:
>> apictl install api-operator \
            --registry-type=HTTP \
            --repository=<REPOSITORY> \
            --username=<USER_NAME> \
            --password=<PASSWORD>

Example:
>> apictl install api-operator \
            --registry-type=HTTP \
            --repository=10.100.5.225:5000/wso2 \
            --username=jennifer \
            --password=********
```

### 1.6. HTTPS Private Registry

Registry type: **HTTPS**

```sh
Format:
>> apictl install api-operator \
            --registry-type=HTTPS \
            --repository=<REPOSITORY> \
            --username=<USER_NAME> \
            --password=<PASSWORD>

Example:
>> apictl install api-operator --registry-type=HTTPS --repository=10.100.5.225:5000/wso2 --username=jennifer --password=********
```

### 1.7. QUAY.IO Registry

Registry type: **QUAY**

```sh
Format:
>> apictl install api-operator \
            --registry-type=QUAY \
            --repository=<REPOSITORY> \
            --username=<USER_NAME> \
            --password=<PASSWORD>

Example:
>> apictl install api-operator \
            --registry-type=QUAY \
            --repository=john \
            --username=john \
            --password=********
```

## 2. Installation Configurations

### 2.1. Install with Default Configurations

You can quick start and try API Operator with the default configurations that we set for you.

```sh
>> apictl install api-operator \
            --registry-type=DOCKER_HUB \
            --repository=docker.io/wso2 \
            --username=john \
            --password=*******
```

#### 2.1.1. What are the default Configurations?

You can find the default configurations of the API Operator in the extracted `k8s-api-operator-<VERSION>.zip` file in
the [releases](https://github.com/wso2/k8s-api-operator/releases).

Default configurations: `K8S-API-OPERATOR-HOME/api-operator/controller-artifacts/`

### 2.2. Install with Customized Configurations and Offline Installation

You can specify your default configurations for the API Operator by specifying the configuration file, directory with
config files or URL.

Change the configurations files with your defaults and specify the location with the flag `-f` or `--from-file`.
This `<ALL_CONFIGURATIONS>` can be a **file**, **directory** or an **URL**

```sh
>> apictl install api-operator \
            --registry-type=DOCKER_HUB \
            --repository=docker.io/wso2 \
            --username=john \
            --password=*******
            -f <ALL_CONFIGURATIONS>
```