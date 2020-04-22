package analytics

import (
	"errors"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/cert"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/mgw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logger = log.Log.WithName("analytics")

// k8s configs
const (
	wso2NameSpaceConst   = "wso2-system"
	analyticsConfName    = "analytics-config"
	analyticsSecretConst = "analyticsSecret"

	usernameConst = "username"
	passwordConst = "password"
	certConst     = "cert_security"
)

// MGW configs
const (
	analyticsEnabledConst          = "analyticsEnabled"
	uploadingTimeSpanInMillisConst = "uploadingTimeSpanInMillis"
	rotatingPeriodConst            = "rotatingPeriod"
	uploadFilesConst               = "uploadFiles"
	hostnameConst                  = "hostname"
	portConst                      = "port"
)

func Handle(client *client.Client, userNamespace string) error {
	analyticsConf := k8s.NewConfMap()
	errConf := k8s.Get(client, types.NamespacedName{Namespace: wso2NameSpaceConst, Name: analyticsConfName}, analyticsConf)
	if errConf != nil {
		logger.Info("Disabling analytics since the analytics configuration related config map not found")
		mgw.Configs.AnalyticsEnabled = false
	} else {
		if analyticsConf.Data[analyticsEnabledConst] == "true" {
			// gets the data from analytics secret
			analyticsSecret := k8s.NewSecret()
			errSecret := k8s.Get(client, types.NamespacedName{
				Namespace: wso2NameSpaceConst,
				Name:      analyticsConf.Data[analyticsSecretConst],
			}, analyticsSecret)

			if errSecret == nil && isValidSecret(analyticsSecret) {
				analyticsCertSecretName := string(analyticsSecret.Data[certConst])
				analyticsCertSecret := k8s.NewSecret()
				// checks if the certificate exists in the namespace of the API
				errCertNs := k8s.Get(client, types.NamespacedName{Name: analyticsCertSecretName, Namespace: userNamespace}, analyticsCertSecret)
				if errCertNs != nil {
					logger.Info("Analytics certificate is not found in the user namespace. Finding it in system namespace", "user_namespace", userNamespace, "system_namespace", wso2NameSpaceConst)
					errCopyCert := k8s.Get(client, types.NamespacedName{Name: analyticsCertSecretName, Namespace: wso2NameSpaceConst}, analyticsCertSecret)
					if errCopyCert != nil {
						logger.Error(errCopyCert, "Error getting analytics certificate in the system namespace", "system_namespace", wso2NameSpaceConst)
						return errCopyCert
					}
					// copy to user namespace
					analyticsCertSecret.SetResourceVersion("")
					errCopyCert = k8s.Create(client, analyticsCertSecret)
					if errCopyCert != nil {
						return errCopyCert
					}
				}
				// Configure MGW and add cert
				setMgwConfigs(analyticsConf, analyticsSecret)
				cert.Add(analyticsCertSecret, "analytics")
			} else {
				if errSecret == nil {
					errSecret = errors.New("required field in the secret is missing the secret: " + analyticsConf.Data[analyticsSecretConst])
				}
				logger.Error(errSecret, "Error in analytics secret", "secret", analyticsSecret)
			}

		} else {
			logger.Info("Analytics is disabled in the configuration")
		}
	}

	return nil
}

func isValidSecret(secret *corev1.Secret) bool {
	return secret.Data != nil && secret.Data[usernameConst] != nil &&
		secret.Data[passwordConst] != nil && secret.Data[certConst] != nil
}

// setMgwConfigs enable analytics and set MGW configs
func setMgwConfigs(confMap *corev1.ConfigMap, secret *corev1.Secret) {
	mgw.Configs.AnalyticsEnabled = true

	mgw.Configs.UploadingTimeSpanInMillis = confMap.Data[uploadingTimeSpanInMillisConst]
	mgw.Configs.RotatingPeriod = confMap.Data[rotatingPeriodConst]
	mgw.Configs.UploadFiles = confMap.Data[uploadFilesConst]
	mgw.Configs.AnalyticsHostname = confMap.Data[hostnameConst]
	mgw.Configs.AnalyticsPort = confMap.Data[portConst]
	mgw.Configs.AnalyticsUsername = string(secret.Data[usernameConst])
	mgw.Configs.AnalyticsPassword = string(secret.Data[passwordConst])
}
