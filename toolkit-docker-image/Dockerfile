#
# Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
#
# WSO2 Inc. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
# 
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

FROM adoptopenjdk/openjdk8:jdk8u222-b10-alpine

#copy java and micro-gw tool kit
# COPY ./files/jdk1.8.0_* /java
ADD ./files/wso2am-micro-gw-toolkit-linux-3.1.0-beta /wso2am-micro-gw-toolkit-linux-3.1.0-beta

# set environment variables
# ENV JAVA_HOME=/java
# ENV PATH=$JAVA_HOME/bin:$PATH

ENV BAL_HOME=/wso2am-micro-gw-toolkit-linux-3.1.0-beta
ENV PATH=$BAL_HOME/bin:$PATH






