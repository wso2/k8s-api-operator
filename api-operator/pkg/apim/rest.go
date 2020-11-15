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
