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
	analyticsEnabled  string = "false"
	analyticsUsername string = "admin"
	analyticsPassword string = "admin"

	//throttlingConfig
	enabledGlobalTMEventPublishing string = "false"
)
