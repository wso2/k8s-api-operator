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

package restserver

import (
	"crypto/tls"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jessevdk/go-flags"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/server/api/models"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/server/api/restserver/operations"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/server/api/restserver/operations/a_p_is_all"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logRestServer = logf.Log.WithName("server.operator")

//go:generate swagger generate server --target ../../api --name Restapi --spec ../../resources/apiSwagger.yaml --server-package restserver --principal models.Principal

func configureFlags(api *operations.RestapiAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.RestapiAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.ApplicationZipProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		dataZip, err := ioutil.ReadFile(os.TempDir() + "/done.zip")
		if err != nil {
			logRestServer.Error(err, "Error when reading zipped file")
			return err
		}
		if dataZip != nil {
			_, err = w.Write(dataZip)
			if err != nil {
				logRestServer.Error(err, "Error when writing zipped file to response")
				return err
			}
		}
		return nil
	})

	// Applies when the Authorization header is set with the Basic scheme
	api.BasicAuthAuth = func(user string, pass string) (*models.Principal, error) {
		if user != "admin" || pass != "admin" {
			return nil, errors.New(401, "Credentials are invalid")
		}

		p := models.Principal{
			Token:    "xxxx",
			Tenant:   "xxxx",
			Username: user,
		}
		return &p, nil
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()
	if api.ApIsAllGetApisHandler == nil {
		api.ApIsAllGetApisHandler = a_p_is_all.GetApisHandlerFunc(func(params a_p_is_all.GetApisParams,
			principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation a_p_is_all.GetApis has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.

	tlsConfig.Certificates, _ = getCertificates("/home/wso2/security/tls.crt", "/home/wso2/security/tls.key")
}

func getCertificates(publicKeyPath, privateKeyPath string) ([]tls.Certificate, error) {
	certificates := make([]tls.Certificate, 1)
	tlsCertificate := publicKeyPath
	tlsCertificateKey := privateKeyPath
	certificate, err := tls.LoadX509KeyPair(tlsCertificate, tlsCertificateKey)
	if err != nil {
		logRestServer.Error(err, "Error while loading key pair")
		return nil, err
	}
	certificates[0] = certificate
	return certificates, nil
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}

func StartRestServer() {
	swaggerSpec, err := loads.Embedded(SwaggerJSON, FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewRestapiAPI(swaggerSpec)
	server := NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Internal Utility API"
	parser.LongDescription = "This document specifies a **RESTful API** for allowing you to access internal data " +
		".\nPlease see [full swagger definition](https://raw.githubusercontent.com/wso2/carbon-apimgt/master/components/apimgt/org.wso2.carbon.apimgt.internal.service/src/main/resources/api.yaml) " +
		"of the API which is written using [swagger 2.0](http://swagger.io/) specification.\n"
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	server.ConfigureAPI()
	server.TLSHost = "0.0.0.0"
	server.TLSPort = 9445

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
