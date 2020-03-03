// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import ballerina/http;

//This function check if the response is a json and modify the json response.
public function validateResponse (http:Caller outboundEp, http:Response res) {
//Client only supports json. Therefore we need to make sure only json responses are returned
    var payload = res.getJsonPayload();
    if (payload is json) {
        int length = payload.toJsonString().length();
        json resJson = {
            "pets": payload,
            "length": length
        };
        res.setJsonPayload(<@untainted> resJson);
    }
}