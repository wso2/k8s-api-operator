#Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

#This image will be used by Kaniko pod to create the micro-gw image

# Custom image with java and toolkit has been build for stage1
FROM wso2am/wso2micro-gw-toolkit:3.0.0-beta2 as toolkit
#init the micro-gw project
RUN micro-gw init project
#copy swagger definitions to the project location
RUN cp /usr/wso2/swagger/* ./project/api_definitions/
#generate the executable
RUN micro-gw build project
#copy the executable to the home/exec location
RUN mkdir -p /home/exec

RUN cp ./project/target/* /home/exec/

#Stage 2: Generate balx will be passed to the micro-gw runtime
FROM wso2/wso2micro-gw:3.0.0-beta2
#change permission to create and copy the executable
USER root
RUN  mkdir -p /home/exec
COPY --from=toolkit /home/exec /home/exec

RUN keytool -list -storepass ballerina -v -keystore /ballerina/runtime/bre/security/ballerinaTruststore.p12

USER ballerina
#project name has to be passed to the ballerina image as a env variable
ENV project="project"