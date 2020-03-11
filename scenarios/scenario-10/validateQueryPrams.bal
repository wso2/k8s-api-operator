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

// This function will validate if the request has 'status' query parameter.
public function validateRequest (http:Caller outboundEp, http:Request req) {
    map<string[]> params = req.getQueryParams();
    string[]? paramStatus = params["status"];
    if (paramStatus is ()) {
        http:Response res = new;
        res.statusCode = 400;
        json message = {"error": "Missing a required parameter"};
        res.setPayload(message);
        var status = outboundEp->respond(res);
    } else if (paramStatus[0] != "available" && paramStatus[0] != "pending" && paramStatus[0] != "sold" && paramStatus[0] != "soon") {
        http:Response res = new;
        res.statusCode = 422;
        json message = {"error": "Invalid status parameter"};
        res.setPayload(message);
        var status = outboundEp->respond(res);
    }
}