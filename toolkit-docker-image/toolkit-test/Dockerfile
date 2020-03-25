FROM renukafernando/wso2micro-gw-toolkit:3.1.0-rc2-2 as toolkit

#init the micro-gw project
RUN micro-gw init project

#copy swagger definitions (seperated in projects directories) to the project location
COPY swagger.json ./project/api_definitions/

#generate the executable
RUN micro-gw build project

#copy the executable to the home/exec location
RUN mkdir -p /home/exec

RUN cp ./project/target/*.jar /home/exec/

#Stage 2: Generated jar will be passed to the micro-gw runtime
FROM wso2/wso2micro-gw:3.1.0-rc2

#change permission to create and copy the executable
USER root

RUN  mkdir -p /home/exec

COPY --from=toolkit /home/exec /home/exec

#copy microgateway conf
RUN cp /usr/wso2/mgwconf/* /home/ballerina/conf


#RUN echo yes | /home/ballerina/wso2/lib/jdk8*/bin/keytool -import  -storepass ballerina -keystore /home/ballerina/wso2/runtime/bre/security/ballerinaTruststore.p12 -alias "wso2am310-secretalias" -file /usr/wso2/certs/wso2am310-secret/server.pem
#RUN echo yes | /home/ballerina/wso2/lib/jdk8*/bin/keytool -import  -storepass ballerina -keystore /home/ballerina/wso2/runtime/bre/security/ballerinaTruststore.p12 -alias "wso2analytics260" -file /usr/wso2/analyticssecret/wso2analytics.pem


RUN chmod a+w /home/ballerina
USER ballerina
#project name has to be passed to the ballerina image as a env variable
ENV project="project"