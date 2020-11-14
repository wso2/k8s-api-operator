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

package security

import (
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logDef = log.Log.WithName("security.default")

func Default(client *client.Client, apiNamespace string, owner *[]metav1.OwnerReference) (*[]mgw.JwtTokenConfig, error) {
	var defaultSecConfArray []mgw.JwtTokenConfig
	//copy default sec in wso2-system to user namespace
	securityDefault := &wso2v1alpha1.Security{}
	//check default security already exist in user namespace
	errGetSec := k8s.Get(client, types.NamespacedName{Name: defaultSecurity, Namespace: apiNamespace}, securityDefault)
	if errGetSec != nil && errors.IsNotFound(errGetSec) {
		logDef.Info("Get default-security", "from namespace", config.SystemNamespace)
		//retrieve default-security from wso2-system namespace
		errSec := k8s.Get(client, types.NamespacedName{Name: defaultSecurity, Namespace: config.SystemNamespace}, securityDefault)
		if errSec != nil {
			logDef.Error(errSec, "Error getting default security", "namespace", config.SystemNamespace)
			return nil, errSec
		}
		for _, defaultSecurityConf := range securityDefault.Spec.SecurityConfig {
			defaultSecConf := mgw.JwtTokenConfig{}
			var defaultCert = k8s.NewSecret()
			//check default certificate exists in user namespace
			err := k8s.Get(client, types.NamespacedName{Name: defaultSecurityConf.Certificate, Namespace: apiNamespace}, defaultCert)
			if err != nil && errors.IsNotFound(err) {
				errCert := k8s.Get(client, types.NamespacedName{Name: defaultSecurityConf.Certificate, Namespace: config.SystemNamespace}, defaultCert)
				if errCert != nil {
					return nil, errCert
				}
				//copying default cert as a secret to user namespace
				var defaultCertName string
				var defaultCertValue []byte
				for cert, value := range defaultCert.Data {
					defaultCertName = cert
					defaultCertValue = value
				}
				defCertData := map[string][]byte{defaultCertName: defaultCertValue}
				newDefaultSecret := k8s.NewSecretWith(types.NamespacedName{Namespace: apiNamespace, Name: defaultSecurityConf.Certificate}, &defCertData, nil, owner)
				errCreateSec := k8s.Apply(client, newDefaultSecret)
				if errCreateSec != nil {
					return nil, errCreateSec
				} else {
					//mount certs
					alias := cert.Add(newDefaultSecret, "security")
					defaultSecConf.CertificateAlias = alias
				}
			} else if err != nil {
				logDef.Error(err, "Error getting default certificate", "from namespace", apiNamespace)
				return nil, err
			} else {
				//mount certs
				alias := cert.Add(defaultCert, "security")
				defaultSecConf.CertificateAlias = alias
			}
			if defaultSecurityConf.Issuer != "" {
				defaultSecConf.Issuer = defaultSecurityConf.Issuer
			}
			if defaultSecurityConf.Audience != "" {
				defaultSecConf.Audience = defaultSecurityConf.Audience
			}
			defaultSecConf.ValidateSubscription = defaultSecurityConf.ValidateSubscription
			// append JwtConfigs
			defaultSecConfArray = append(defaultSecConfArray, defaultSecConf)
		}
		//copying default security to user namespace
		logDef.Info("copying default security to " + apiNamespace)
		newDefaultSecurity := copyDefaultSecurity(securityDefault, apiNamespace, *owner)
		errCreateSecurity := k8s.Create(client, newDefaultSecurity)
		if errCreateSecurity != nil {
			logDef.Error(errCreateSecurity, "error creating secret for default security in user namespace")
			return nil, errCreateSecurity
		}
		logDef.Info("default security successfully copied to " + apiNamespace + " namespace")

	} else if errGetSec != nil {
		logDef.Error(errGetSec, "error getting default security from user namespace")
		return nil, errGetSec
	} else {
		logDef.Info("Default security exists in the namespace", "namespace", apiNamespace)
		// check default cert exist in api namespace
		for _, securityDefaultConf := range securityDefault.Spec.SecurityConfig {
			defaultSecConf := mgw.JwtTokenConfig{}
			var defaultCertUsrNs = k8s.NewSecret()
			err := k8s.Get(client, types.NamespacedName{Name: securityDefaultConf.Certificate, Namespace: apiNamespace}, defaultCertUsrNs)
			if err != nil {
				return nil, err
			} else {
				//mount certs
				alias := cert.Add(defaultCertUsrNs, "security")
				defaultSecConf.CertificateAlias = alias
				defaultSecConf.ValidateSubscription = securityDefaultConf.ValidateSubscription
			}
			if securityDefaultConf.Issuer != "" {
				defaultSecConf.Issuer = securityDefaultConf.Issuer
			}
			if securityDefaultConf.Audience != "" {
				defaultSecConf.Audience = securityDefaultConf.Audience
			}
			// append JwtConfigs
			defaultSecConfArray = append(defaultSecConfArray, defaultSecConf)
		}
	}
	return &defaultSecConfArray, nil
}

func copyDefaultSecurity(securityDefault *wso2v1alpha1.Security, userNameSpace string, owner []metav1.OwnerReference) *wso2v1alpha1.Security {
	securityConf := wso2v1alpha1.SecurityConfig{}
	var securityConfArray []wso2v1alpha1.SecurityConfig
	for _, securityDefaultConf := range securityDefault.Spec.SecurityConfig {
		securityConf = wso2v1alpha1.SecurityConfig{
			Certificate: securityDefaultConf.Certificate,
			Audience:    securityDefaultConf.Audience,
			Issuer:      securityDefaultConf.Issuer,
		}
		securityConfArray = append(securityConfArray, securityConf)
	}

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
