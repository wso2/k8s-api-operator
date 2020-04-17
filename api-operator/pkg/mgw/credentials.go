package mgw

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var logger = log.Log.WithName("mgw")

func SetCredentials(client *client.Client, securityType string, namespacedName types.NamespacedName) error {
	sha1Hash := sha1.New()
	var userName string
	var password []byte

	//get the secret included credentials
	credentialSecret := k8s.NewSecret()
	err := k8s.Get(client, namespacedName, credentialSecret)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	//get the username and the password
	for k, v := range credentialSecret.Data {
		if strings.EqualFold(k, "username") {
			userName = string(v)
		}
		if strings.EqualFold(k, "password") {
			password = v
		}

	}
	if securityType == "Basic" {

		Configs.BasicUsername = userName
		_, err := sha1Hash.Write([]byte(password))
		if err != nil {
			logger.Info("error in encoding password")
			return err
		}
		//convert encoded password to a hex string
		Configs.BasicPassword = hex.EncodeToString(sha1Hash.Sum(nil))
	}
	if securityType == "Oauth" {
		Configs.KeymanagerUsername = userName
		Configs.KeymanagerPassword = string(password)
	}
	return nil
}
