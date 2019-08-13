Creating MGW-ToolKit Image
Note that the root directory is called <dockerfile-home>

1. Create a directory called "files" in <dockerfile-home>
2. Place relevant Java distribution and MGW toolkit distribution in the "files" directory (unzipped)
3. Change the mgw toolkit version in the Dockerfile.
4. Place the Dockerfile in <dockerfile-home>
5. Execute the following command to build the Dockerfile
v       docker build -t wso2am/<toolkit name and version for the image name> .
        eg: docker build -t wso2am/wso2am-micro-gw-toolkit-3.0.1 .
6. Pushing the built dockerimage to docker registry 
        docker push wso2am/<toolkit name and version for the image name>
        docker push wso2am/wso2am-micro-gw-toolkit-3.0.1