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

package api

var (

	//listenerConfig
	keystorePath       string = "${ballerina.home}/bre/security/ballerinaKeystore.p12"
	keystorePassword   string = "ballerina"
	truststorePath     string = "${ballerina.home}/bre/security/ballerinaTruststore.p12"
	truststorePassword string = "ballerina"

	//keymanager
	keymanagerServerurl string = "https://localhost:9443"
	keymanagerUsername  string = "admin"
	keymanagerPassword  string = "admin"

	//jwtTokenConfig
	issuer           string = "https://localhost:9443/oauth2/token"
	audience         string = "http://org.wso2.apimgt/gateway"
	certificateAlias string = "wso2apim"

	//analytics
	analyticsEnabled          string = "false"
	analyticsUsername         string = "admin"
	analyticsPassword         string = "admin"
	uploadingTimeSpanInMillis string = "600000"
	rotatingPeriod            string = "600000"
	uploadFiles               string = "true"
	verifyHostname            string = "true"
	hostname                  string = "localhost"
	port                      string = "9444"

	//throttlingConfig
	enabledGlobalTMEventPublishing string = "false"

	//basic authentication
	basicUsername string = "generalUser1"
	basicPassword string = "5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8"
)
