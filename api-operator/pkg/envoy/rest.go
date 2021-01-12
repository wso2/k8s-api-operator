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

package envoy

import (
	"crypto/tls"
	"gopkg.in/resty.v1"
	"os"
	"time"
)

func invokePOSTRequestWithBytes(url string, headers map[string]string, body []byte) (*resty.Response, error) {
	if insecureDeploy {
		resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) // To bypass errors in SSL certificates
	} else {
		resty.SetTLSClientConfig(&tls.Config{RootCAs: certPool, InsecureSkipVerify: false})
	}

	if os.Getenv("HTTP_PROXY") != "" {
		resty.SetProxy(os.Getenv("HTTP_PROXY"))
	} else if os.Getenv("HTTPS_PROXY") != "" {
		resty.SetProxy(os.Getenv("HTTPS_PROXY"))
	} else if os.Getenv("http_proxy") != "" {
		resty.SetProxy(os.Getenv("http_proxy"))
	} else if os.Getenv("https_proxy") != "" {
		resty.SetProxy(os.Getenv("https_proxy"))
	}
	resty.SetTimeout(time.Duration(DefaultHttpRequestTimeout) * time.Millisecond)
	resp, err := resty.R().SetHeaders(headers).SetBody(body).Post(url)
	return resp, err
}
