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
    >> operator-sdk build wso2/k8s-api-operator:2.0.0-rc1
    ```

1.  Replace the image name in deploy/controller-artifacts/operator.yaml#L36

1.  Push it to a registry:
    ```sh
    >> docker push wso2/k8s-api-operator:2.0.0-rc1
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

Create following certs in the directory `/home/wso2/security/` (or update the cert directory in
[this file](pkg/envoy/server/api/restserver/configure_restapi.go) with a directory of your choice)

The content of the following file is the content of the secret [operator-secret](deploy/controller-artifacts/operator-secret.yaml).

- Operator Private Key File
    ```pem
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEAscJkNq9mdY9tnPil0CXmvxyvSmaBMmkfC3yQYtm5DtB8Z+6O
    0z74uB7qY/oaLE9GBjgGKc95oNZ4AokjUDqAevCumvmHY/oXMPQ+75KSFECWu9GF
    GQkIRcGtfXlCL5OZkbnocZJ+dSn3T+iiNK3d4xck1M7MqFD/rlICChYDSz/tAd9w
    VzrD0uOynqvVChpSyHFXezxhoXsRm6BzXO2wFc4vI7hF0fYZS2nj0zHGanhhna3I
    lupWALrLt7TyMHz7VsC1Kp6y4Poh3jz174UmYctO9ajQc7g9J2rwvaK6f0gWYTqP
    O5bxhFMjBbEwYTIu91eT69+BbxL6Q05/AdRNmQIDAQABAoIBAAxxu2pIpTedruLK
    VXFY9epzK6JdwrWwvkejlMgWzenHlq3/+We4hNj+8RHGtIZdll1hFq1epPkpioVC
    7IH1VoFE0DRxO5U4MN9weedzr15TlznW1SaHh1i60lZyYrvJ7XpSNX+d7dRt965U
    buaZNWtsd1uejp1J7lxZyWLDX7+owB7WYhTCoCZk7/But6JBG7EPmNblUGj0E2YI
    JafpC5cQxkZsZO/Fml5KvbzEUxzvuLTuecTkmRBvdx6XRneKbh42z4H41uNCaR6A
    lWxev0DgrhcTHbex018aIcOFw6GPOBJBoNHXBepUeNcFmkDI7bJ4j14ZK5f91f4X
    gEH3H8ECgYEA3cNCtSyIVqzvzNIu281Wa2FsTA92igMhHTL4GG5TzELtkaGEX4lZ
    ExppaYSGZgdohc4HL3LyRmdWKql6vzE5XokgNsmrdakEG9zHDtxriAjTP3pkZ6wk
    GG1hHNymF0Of3XoZckO1btNFVBYcRne5vUFAvb4Ls9UvTNqlzuEqHc0CgYEAzTP7
    iThFrnx9eevPU8x1dI+uHgepwkn1RqYPfKFzFSSLgTVGruxLN4XApL20FnC92xrp
    Z6CZw+E7I/Q0HTDDL2daifDjkVsnGrYmc2iSm+nS2F32P44+fYFDlfZwhVz6Ctfb
    h3XE9Kqr37QTnXKOLavrb1zkAGyAeEsdSSeeQv0CgYBqAO8/KTVWsT3DW2j4unOn
    yp740J9qI0rN6VI8Y2h9CDUFWv6qqD3C6uoefTG9TadB5pT6smhrDPRcWj0JbV8t
    +EBE0Cu8h3kmVGd2jBh+ozFPc5LRF7D9WDOGl1ZxYmrldHr7arAsdKL8KcGEUbCg
    bbOjv1das/nzM8T0Wh9GtQKBgQCp6MX4++AutyPKZSfZki0bI7Efamb02fo+0ld2
    cdRxiD3+8ZciVcN+KMC3Z+CKyDVcC++Bf6hyWbd3cMgJ94tWX/TGzPARNnGtm29B
    FjB26uhLgZnZTDWQBA1rSZAnzTG48rzyb+ByWjNQWrH3J5h0VqruHfMoKq9Ba6jH
    HwfbHQKBgC8NN1D6vCgHKcbm7IWOYem05LBHMlFFQqk0z7kkNlI4CDeJuDwQEGaB
    ZsbwmOUW+rjodP3bYUVlNxCaGY+RSMfcIWBINWOwpFeiWnQ9bNRHcui5awXwmpzo
    Rn2esOHA+y7T8g0o2Yvkc7rHkAAdjMNiLHUpQ1MwAUNiB9LLjZEu
    -----END RSA PRIVATE KEY-----
    ```

- Operator Public Cert File
    ```pem
    -----BEGIN CERTIFICATE-----
    MIIDqTCCApGgAwIBAgIUfIW77suf7AiEzZ+bRJk1vZ8l3bkwDQYJKoZIhvcNAQEL
    BQAwZDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRYwFAYDVQQHDA1Nb3VudGFp
    biBWaWV3MQ0wCwYDVQQKDARXU08yMQ0wCwYDVQQLDARXU08yMRIwEAYDVQQDDAls
    b2NhbGhvc3QwHhcNMjAxMjA5MDQwMTI2WhcNMzAxMjA3MDQwMTI2WjBkMQswCQYD
    VQQGEwJVUzELMAkGA1UECAwCQ0ExFjAUBgNVBAcMDU1vdW50YWluIFZpZXcxDTAL
    BgNVBAoMBFdTTzIxDTALBgNVBAsMBFdTTzIxEjAQBgNVBAMMCWxvY2FsaG9zdDCC
    ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALHCZDavZnWPbZz4pdAl5r8c
    r0pmgTJpHwt8kGLZuQ7QfGfujtM++Lge6mP6GixPRgY4BinPeaDWeAKJI1A6gHrw
    rpr5h2P6FzD0Pu+SkhRAlrvRhRkJCEXBrX15Qi+TmZG56HGSfnUp90/oojSt3eMX
    JNTOzKhQ/65SAgoWA0s/7QHfcFc6w9Ljsp6r1QoaUshxV3s8YaF7EZugc1ztsBXO
    LyO4RdH2GUtp49Mxxmp4YZ2tyJbqVgC6y7e08jB8+1bAtSqesuD6Id489e+FJmHL
    TvWo0HO4PSdq8L2iun9IFmE6jzuW8YRTIwWxMGEyLvdXk+vfgW8S+kNOfwHUTZkC
    AwEAAaNTMFEwHQYDVR0OBBYEFDbf7A7HV9eSUMKidPGHvyJW6hrzMB8GA1UdIwQY
    MBaAFDbf7A7HV9eSUMKidPGHvyJW6hrzMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
    hvcNAQELBQADggEBAAuTlShWOx7bN1Jhps6wRgU6M8cvmeZ+LLkS2AwZ75sKNpT0
    qxXih3UkqAZCOIEs9wCZKRnRJuBSOK55h4xF4KNqexTFkbgxOMncS+Mxg69s2nIg
    rn8fEHrWftua3vsnMdNfk3AjCXflRJb0x5J9vDIQ7qxSVTeHe4gLaa2uzzUvcXqM
    H2BPKnVMGi3epXKHRxojyYLuTmfeFBXAnbFBM+sBDGg1k9qKC3B6jlOVD3fXPEqB
    NlPPkjRGI9pI82NpEIe9TWd7O40sd2M6dMhb8eyTtohnEpZDZ1IdCI904rVUSEip
    jU0NqG9zgErbLF7LhrtaI2YGBBp5LVlRxCtGeRk=
    -----END CERTIFICATE-----
    ```

Install API Operator from distribution and delete the API Operator deployment.

```sh
kubectl delete deploy api-operator
```

Execute `cmd/manager/main.go` file in debug mode.
