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
