/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

Creating MGW-ToolKit Image
Note that the root directory is called <dockerfile-home>

1. Create a directory called "files" in <dockerfile-home>
2. Place relevant Java distribution and MGW toolkit distribution in the "files" directory (unzipped)
3. Change the mgw toolkit version in the Dockerfile.
4. Place the Dockerfile in <dockerfile-home>
5. Execute the following command to build the Dockerfile
        docker build -t wso2am/<toolkit name and version for the image name> .
        eg: docker build -t wso2am/wso2am-micro-gw-toolkit-3.0.1 .
6. Pushing the built dockerimage to docker registry 
        docker push wso2am/<toolkit name and version for the image name>
        docker push wso2am/wso2am-micro-gw-toolkit-3.0.1