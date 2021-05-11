// Copyright (c) WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package endpoints

import (
	"fmt"
	wso2v1alpha1 "github.com/wso2/k8s-api-operator/api-operator/pkg/apis/wso2/v1alpha1"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/kaniko"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	aliasSecretKey       = "alias"
	certificateSecretKey = "certificate.crt"
)

var loggerCert = log.Log.WithName("endpoints.cert")

var (
	allAlias = make(map[string]struct{})
	void     = struct{}{}
)

// HandleCerts handles the endpoint certificates defined with the API-CTL (API Controller) project
func HandleCerts(client *client.Client, kanikoProps *kaniko.JobProperties, api *wso2v1alpha1.API) error {
	for _, certSecretName := range api.Spec.Definition.EndpointCertificates {
		secret := &corev1.Secret{}
		if err := k8s.Get(client, types.NamespacedName{Namespace: api.Namespace, Name: certSecretName}, secret); err != nil {
			loggerCert.Error(err, "Error reading endpoint certificate")
			return err
		}
		if err := addCert(kanikoProps, secret); err != nil {
			loggerCert.Error(err, "Error adding endpoint certificate")
			return err
		}
		loggerCert.Info("Added endpoint cert for API", "api", api, "secret", certSecretName)
	}
	return nil
}

func addCert(kanikoProps *kaniko.JobProperties, certSecret *corev1.Secret) error {
	if err := validateSecret(certSecret); err != nil {
		loggerCert.Error(err, "Invalid endpoint secret", "secret", certSecret,
			"namespace", certSecret.Namespace)
		return err
	}

	alias := string(certSecret.Data[aliasSecretKey])

	// skip adding cert if alias already exists
	if _, ok := allAlias[alias]; ok {
		loggerCert.Info("Alias of endpoint certificate already exists. Skip importing certificate.",
			"alias", alias, "secret", certSecret.Name, "namespace", certSecret.Namespace)
	}

	allAlias[alias] = void
	cert.Add(kanikoProps, alias, certSecret.Name, certificateSecretKey)
	return nil
}

func validateSecret(certSecret *corev1.Secret) error {
	requiredKeys := []string{aliasSecretKey, certificateSecretKey}
	for _, key := range requiredKeys {
		if _, ok := certSecret.Data[key]; !ok {
			return fmt.Errorf("required key of the sercret %v", certSecret.Name)
		}
	}
	return nil
}
