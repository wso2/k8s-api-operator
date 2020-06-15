# Creating MGW-ToolKit Image

1. Change the mgw toolkit version (i.e. ARG `WSO2_SERVER_VERSION`) in the Dockerfile.
    ```dockerfile
    ARG WSO2_SERVER_VERSION=3.2.0
    ```

1. Execute the following command to build the docker image from Dockerfile.
    ```sh
    >> docker build -t wso2am/wso2micro-gw-toolkit:<TOOLKIT_VERSION> .
    
    # example:
    >> docker build -t wso2am/wso2micro-gw-toolkit:3.2.0 .
    ```

1. Push the built docker image to docker registry.
    ```sh
    >> docker push wso2am/wso2micro-gw-toolkit:<TOOLKIT_VERSION>
    
    # example:
    >> docker push wso2am/wso2micro-gw-toolkit:3.2.0
    ```
