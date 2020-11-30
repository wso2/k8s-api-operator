package envoy

const (
	mgBasePath                 = "/mgw/1.0"
	mgDeployResourcePath       = "/import/api"
	envoyMgwConfName           = "envoy-mgw-configs"
	envoyMgwSecretName         = "envoymgw-adapter-secret"
	mgwAdapterHostConst        = "mgwAdapterHost"
	mgwInsecureSkipVerifyConst = "mgwInsecureSkipVerify"
	HeaderAuthorization        = "Authorization"
	HeaderAccept               = "Accept"
	HeaderContentType          = "Content-Type"
	HeaderConnection           = "Connection"
	HeaderValueAuthBasicPrefix = "Basic"
	HeaderValueKeepAlive       = "keep-alive"
	DefaultHttpRequestTimeout  = 10000
)
