package security

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logger = log.Log.WithName("security")

type securitySchemeStruct struct {
	SecurityType string             `json:"type"`
	Scheme       string             `json:"scheme,omitempty"`
	Flows        *authorizationCode `json:"flows,omitempty"`
}

type authorizationCode struct {
	AuthorizationCode scopeSet `json:"authorizationCode,omitempty"`
}

type scopeSet struct {
	AuthorizationUrl string            `json:"authorizationUrl"`
	TokenUrl         string            `json:"tokenUrl"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

func Handle(client *client.Client, securityMap map[string][]string, userNameSpace string, secSchemeDefined bool) (map[string]securitySchemeStruct, *[]mgw.JwtTokenConfig, error) {
	var securityDefinition = make(map[string]securitySchemeStruct)
	//to add multiple certs with alias

	jwtConfArray := []mgw.JwtTokenConfig{}
	securityInstance := &wso2v1alpha1.Security{}
	var certificateSecret = k8s.NewSecret()
	for secName, scopeList := range securityMap {
		//retrieve security instances
		errGetSec := k8s.Get(client, types.NamespacedName{Name: secName, Namespace: userNameSpace}, securityInstance)
		if errGetSec != nil && errors.IsNotFound(errGetSec) {
			return securityDefinition, &jwtConfArray, errGetSec
		}
		if strings.EqualFold(securityInstance.Spec.Type, oauthConst) {
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				errc := k8s.Get(client, types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
				if errc != nil && errors.IsNotFound(errc) {
					logger.Info("defined certificate is not found")
					return securityDefinition, &jwtConfArray, errc
				} else {
					logger.Info("defined certificate successfully retrieved")
				}
				//mount certs
				_ = cert.Add(certificateSecret, "security")

				//get the keymanager server URL from the security kind
				mgw.Configs.KeyManagerServerUrl = securityConf.Endpoint
				//fetch credentials from the secret created
				errGetCredentials := mgw.SetCredentials(client, oauthConst, types.NamespacedName{Namespace: userNameSpace, Name: securityConf.Credentials})
				if errGetCredentials != nil {
					logger.Error(errGetCredentials, "Error occurred when retrieving credentials for Oauth")
				} else {
					logger.Info("Credentials successfully retrieved for security " + secName)
				}
				if !secSchemeDefined {
					//add scopes
					scopes := map[string]string{}
					for _, scopeValue := range scopeList {
						scopes[scopeValue] = "grant " + scopeValue + " access"
					}
					//creating security scheme
					scheme := securitySchemeStruct{
						SecurityType: oauth2Type,
						Flows: &authorizationCode{
							scopeSet{
								authorizationUrl,
								tokenUrl,
								scopes,
							},
						},
					}
					securityDefinition[secName] = scheme
				}
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, jwtConst) {
			logger.Info("retrieving data for security type JWT")
			for _, securityConf := range securityInstance.Spec.SecurityConfig {
				jwtConf := mgw.JwtTokenConfig{}
				errc := k8s.Get(client, types.NamespacedName{Name: securityConf.Certificate, Namespace: userNameSpace}, certificateSecret)
				if errc != nil && errors.IsNotFound(errc) {
					logger.Info("defined certificate is not found")
					return securityDefinition, &jwtConfArray, errc
				} else {
					logger.Info("defined certificate successfully retrieved")
				}
				//mount certs
				alias := cert.Add(certificateSecret, "security")
				jwtConf.CertificateAlias = alias
				jwtConf.ValidateSubscription = securityConf.ValidateSubscription

				if securityConf.Issuer != "" {
					jwtConf.Issuer = securityConf.Issuer
				}
				if securityConf.Audience != "" {
					jwtConf.Audience = securityConf.Audience
				}

				logger.Info("certificate issuer", "issuer", jwtConf.Issuer)
				jwtConfArray = append(jwtConfArray, jwtConf)
			}
		}
		if strings.EqualFold(securityInstance.Spec.Type, basicSecurityAndScheme) {
			// "existCert = false" for this scenario and do not change the global "existCert" value
			// i.e. if global "existCert" is true, even though the scenario for this swagger is false keep that value as true

			//fetch credentials from the secret created
			errGetCredentials := mgw.SetCredentials(client, "Basic",
				types.NamespacedName{Namespace: userNameSpace, Name: securityInstance.Spec.SecurityConfig[0].Credentials})
			if errGetCredentials != nil {
				logger.Error(errGetCredentials, "Error occurred when retrieving credentials for Basic")
			} else {
				logger.Info("Credentials successfully retrieved for security " + secName)
			}
			//creating security scheme
			if !secSchemeDefined {
				scheme := securitySchemeStruct{
					SecurityType: basicSecurityType,
					Scheme:       basicSecurityAndScheme,
				}
				securityDefinition[secName] = scheme
			}
		}
	}
	return securityDefinition, &jwtConfArray, nil
}
