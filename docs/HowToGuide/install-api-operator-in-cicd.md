# Install API Operator in CI/CD

`apictl` tool enable you to install api-operator and configure registry in interactive mode and non-interactive mode as well. For CI/CD automation you can use the non-interactive feature based on the registry type.

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

For offline installation use the `--from-file` flag to point the operator configs downloaded from the [release](https://github.com/wso2/k8s-api-operator/releases). The flag `--from-file` can be used with any registry type.

## 1. Docker Hub

```sh
Format:
>> apictl install api-operator --registry-type=DOCKER_HUB --repository=<REPOSITORY> --username=<USER_NAME> --password=<PASSWORD>
```
***Note:*** The flag `--password-stdin` can be used instead of `--password` if you want to read password from standard input

```sh
Docker-Hub:
>> apictl install api-operator --registry-type=DOCKER_HUB --repository=docker.io/wso2 --username=john --password=*******
```

## 2. Amazon ECR

```sh
Format:
>> apictl install api-operator --registry-type=AMAZON_ECR --repository=<REPOSITORY_URI> --key-file=<AWS_CREDENTIALS_FILE>

Example:
>> apictl install api-operator --registry-type=AMAZON_ECR --repository=111222333444.dkr.ecr.us-east-1.amazonaws.com/my-ecr-repo --key-file=/Users/wso2/.aws/credentials
```

## 3. GCR

```sh
Format:
>> apictl install api-operator --registry-type=GCR --key-file=<PATH_TO_GCR_SERVICE_ACCOUNT_KEY_FILE_JSON>

Example:
>> apictl install api-operator --registry-type=GCR --key-file=/path/to/gcr/service/account/key/file.json
```

## 4. HTTP private registry

```sh
Format:
>> apictl install api-operator --registry-type=HTTP --repository=<REPOSITORY> --username=<USER_NAME> --password=<PASSWORD>

Example:
>> apictl install api-operator --registry-type=HTTP --repository=10.100.5.225:5000/wso2 --username=jennifer --password=********
```

## 5. HTTPS private registry

```sh
Format:
>> apictl install api-operator --registry-type=HTTPS --repository=<REPOSITORY> --username=<USER_NAME> --password=<PASSWORD>

Example:
>> apictl install api-operator --registry-type=HTTPS --repository=10.100.5.225:5000/wso2 --username=jennifer --password=********
```

## 6. HTTPS private registry

```sh
Format:
>> apictl install api-operator --registry-type=QUAY --repository=<REPOSITORY> --username=<USER_NAME> --password=<PASSWORD>

Example:
>> apictl install api-operator --registry-type=QUAY --repository=john --username=john --password=********
```