FROM ubuntu:18.04

#copy java and micro-gw tool kit
COPY ./files/jdk1.8.0_* /java
ADD ./files/wso2am-micro-gw-toolkit-3.0.1 /wso2am-micro-gw-toolkit-3.0.1

# set environment variables
ENV JAVA_HOME=/java
ENV PATH=$JAVA_HOME/bin:$PATH

ENV BAL_HOME=/wso2am-micro-gw-toolkit-3.0.1
ENV PATH=$BAL_HOME/bin:$PATH






