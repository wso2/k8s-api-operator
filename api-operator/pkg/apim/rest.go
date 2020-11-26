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

package apim

import (
	"crypto/tls"
	"gopkg.in/resty.v1"
	"time"
)

func invokePOSTRequest(url string, headers map[string]string, body interface{}) (*resty.Response, error) {
	httpClient := resty.New()
	if insecure {
		httpClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	} else {
		httpClient.SetTLSClientConfig(&tls.Config{RootCAs: certPool, InsecureSkipVerify: false})
	}

	httpClient.SetTimeout(time.Duration(DefaultHttpRequestTimeout) * time.Millisecond)

	response, err := httpClient.R().SetHeaders(headers).SetBody(body).Post(url)
	return response, err
}

func invokeGETRequest(url string, headers map[string]string) (*resty.Response, error) {
	httpClient := resty.New()
	if insecure {
		httpClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	} else {
		httpClient.SetTLSClientConfig(&tls.Config{RootCAs: certPool, InsecureSkipVerify: false})
	}
	httpClient.SetTimeout(time.Duration(DefaultHttpRequestTimeout) * time.Millisecond)

	response, err := httpClient.R().SetHeaders(headers).Get(url)
	return response, err
}

func invokePUTRequest(url string, headers map[string]string, body interface{}) (*resty.Response, error) {
	httpClient := resty.New()
	if insecure {
		httpClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	} else {
		httpClient.SetTLSClientConfig(&tls.Config{RootCAs: certPool, InsecureSkipVerify: false})
	}
	httpClient.SetTimeout(time.Duration(DefaultHttpRequestTimeout) * time.Millisecond)

	response, err := httpClient.R().SetHeaders(headers).SetBody(body).Put(url)
	return response, err
}

func invokeDELETERequest(url string, headers map[string]string) (*resty.Response, error) {
	httpClient := resty.New()
	if insecure {
		httpClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	} else {
		httpClient.SetTLSClientConfig(&tls.Config{RootCAs: certPool, InsecureSkipVerify: false})
	}
	httpClient.SetTimeout(time.Duration(DefaultHttpRequestTimeout) * time.Millisecond)

	response, err := httpClient.R().SetHeaders(headers).Delete(url)

	return response, err
}
