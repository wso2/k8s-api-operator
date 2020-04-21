package security

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Default(client *client.Client, apiNamespace string, owner *[]metav1.OwnerReference) error {
	defaultSecConf := mgw.JwtTokenConfig{}
	//copy default sec in wso2-system to user namespace
	securityDefault := &wso2v1alpha1.Security{}
	//check default security already exist in user namespace
	errGetSec := k8s.Get(client, types.NamespacedName{Name: defaultSecurity, Namespace: apiNamespace}, securityDefault)

	if errGetSec != nil && errors.IsNotFound(errGetSec) {
		logger.Info("Get default-security", "from namespace", wso2NameSpaceConst)
		//retrieve default-security from wso2-system namespace
		errSec := k8s.Get(client, types.NamespacedName{Name: defaultSecurity, Namespace: wso2NameSpaceConst}, securityDefault)
		if errSec != nil {
			logger.Error(errSec, "Error getting default security", "namespace", wso2NameSpaceConst)
			return errSec
		}

		var defaultCert = k8s.NewSecret()
		//check default certificate exists in user namespace
		err := k8s.Get(client, types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: apiNamespace}, defaultCert)
		if err != nil && errors.IsNotFound(err) {
			errCert := k8s.Get(client, types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: wso2NameSpaceConst}, defaultCert)
			if errCert != nil {
				return errCert
			}
			//copying default cert as a secret to user namespace
			var defaultCertName string
			var defaultCertValue []byte
			for cert, value := range defaultCert.Data {
				defaultCertName = cert
				defaultCertValue = value
			}
			defCertData := map[string][]byte{defaultCertName: defaultCertValue}
			newDefaultSecret := k8s.NewSecretWith(types.NamespacedName{Namespace: apiNamespace, Name: securityDefault.Spec.SecurityConfig[0].Certificate}, &defCertData, nil, owner)
			errCreateSec := k8s.Create(client, newDefaultSecret)

			if errCreateSec != nil {
				return errCreateSec
			} else {
				//mount certs
				alias := cert.Add(newDefaultSecret, "security")
				defaultSecConf.CertificateAlias = alias
			}
		} else if err != nil {
			logger.Error(err, "Error getting default certificate", "from namespace", apiNamespace)
			return err
		} else {
			//mount certs
			alias := cert.Add(defaultCert, "security")
			defaultSecConf.CertificateAlias = alias
		}
		//copying default security to user namespace
		logger.Info("copying default security to " + apiNamespace)
		newDefaultSecurity := copyDefaultSecurity(securityDefault, apiNamespace, *owner)
		errCreateSecurity := k8s.Create(client, newDefaultSecurity)
		if errCreateSecurity != nil {
			logger.Error(errCreateSecurity, "error creating secret for default security in user namespace")
			return errCreateSecurity
		}
		logger.Info("default security successfully copied to " + apiNamespace + " namespace")
		if newDefaultSecurity.Spec.SecurityConfig[0].Issuer != "" {
			defaultSecConf.Issuer = newDefaultSecurity.Spec.SecurityConfig[0].Issuer
		}
		if newDefaultSecurity.Spec.SecurityConfig[0].Audience != "" {
			defaultSecConf.Audience = newDefaultSecurity.Spec.SecurityConfig[0].Audience
		}
		defaultSecConf.ValidateSubscription = newDefaultSecurity.Spec.SecurityConfig[0].ValidateSubscription
	} else if errGetSec != nil {
		logger.Error(errGetSec, "error getting default security from user namespace")
		return errGetSec
	} else {
		logger.Info("Default security exists in the namespace", "namespace", apiNamespace)
		// check default cert exist in api namespace
		var defaultCertUsrNs = k8s.NewSecret()
		err := k8s.Get(client, types.NamespacedName{Name: securityDefault.Spec.SecurityConfig[0].Certificate, Namespace: apiNamespace}, defaultCertUsrNs)
		if err != nil {
			return err
		} else {
			//mount certs
			alias := cert.Add(defaultCertUsrNs, "security")
			defaultSecConf.CertificateAlias = alias
			defaultSecConf.ValidateSubscription = securityDefault.Spec.SecurityConfig[0].ValidateSubscription
		}
		if securityDefault.Spec.SecurityConfig[0].Issuer != "" {
			defaultSecConf.Issuer = securityDefault.Spec.SecurityConfig[0].Issuer
		}
		if securityDefault.Spec.SecurityConfig[0].Audience != "" {
			defaultSecConf.Audience = securityDefault.Spec.SecurityConfig[0].Audience
		}
	}

	// append JwtConfigs
	*mgw.Configs.JwtConfigs = append(*mgw.Configs.JwtConfigs, defaultSecConf)
	return nil
}

func copyDefaultSecurity(securityDefault *wso2v1alpha1.Security, userNameSpace string, owner []metav1.OwnerReference) *wso2v1alpha1.Security {

	securityConf := wso2v1alpha1.SecurityConfig{
		Certificate: securityDefault.Spec.SecurityConfig[0].Certificate,
		Audience:    securityDefault.Spec.SecurityConfig[0].Audience,
		Issuer:      securityDefault.Spec.SecurityConfig[0].Issuer,
	}

	securityConfArray := []wso2v1alpha1.SecurityConfig{}

	securityConfArray = append(securityConfArray, securityConf)
	return &wso2v1alpha1.Security{
		ObjectMeta: metav1.ObjectMeta{
			Name:            defaultSecurity,
			Namespace:       userNameSpace,
			OwnerReferences: owner,
		},
		Spec: wso2v1alpha1.SecuritySpec{
			Type:           securityDefault.Spec.Type,
			SecurityConfig: securityConfArray,
		},
	}
}
