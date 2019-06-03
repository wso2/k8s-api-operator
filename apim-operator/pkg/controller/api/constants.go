package api

const (
	dockerConfig  = "docker-config"
	mgwDockerFile = "dockerfile-conf"
	swaggerVolume = "swagger-volume"

	swaggerLocation       = "/usr/wso2/swagger/"
	dockerFileLocation    = "/usr/wso2/dockerfile/"
	dockerConfLocation    = "/kaniko/.docker"
	dockerFile            = "docker-file"
	policyyamlFile        = "policy-file"
	policyyamlLocation    = "/usr/wso2/policy/"
	mgwConfFile           = "conf-file"
	mgwConfLocation       = "/usr/wso2/mgwconf/"
	analyticsCertFile     = "analytics-cert"
	analyticsCertLocation = "/usr/wso2/analyticscert/"

	replicasConst       = "replicas"
	mgwToolkitImgConst  = "mgwToolkitImg"
	mgwRuntimeImgConst  = "mgwRuntimeImg"
	kanikoImgConst      = "kanikoImg"
	dockerRegistryConst = "dockerRegistry"
	userNameSpaceConst  = "userNameSpace"

	wso2NameSpaceConst    = "wso2-system"
	policyConfigmap       = "policy-configmap"
	mgwConfSecretConst    = "mgw-secret"
	analyticsSecretConst  = "analytics-secret"
	dockerSecretNameConst = "docker-secret"
	controllerConfName    = "controller-config"

	usernameConst = "username"
	passwordConst = "password"
	certConst     = "cert_security"

	dockerhubRegistryUrl = "https://registry-1.docker.io/"
	defaultSecurity      = "default-security"
	securityExtension    = "x-mgw-security"
)
