package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	port = "8080"
)

var (
	tlscert, tlskey string
)

func main() {

	flag.StringVar(&tlscert, "tlsCertFile", "/etc/certs/cert.pem",
		"File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlskey, "tlsKeyFile", "/etc/certs/key.pem",
		"File containing the x509 private key to --tlsCertFile.")

	flag.Parse()

	certs, err := tls.LoadX509KeyPair(tlscert,tlskey)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%v",port),
		TLSConfig:&tls.Config{Certificates: []tls.Certificate{certs}},
	}

	vs := ValidationServerHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate",vs.serve)
	server.Handler = mux

	// start webhook server in new rountine
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and serve webhook server: %v", err)
			os.Exit(1)
		}
	}()

	glog.Infof("Server running for security kind")
	glog.Infof("Server running listening in port: %s", port)

	// listening shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Info("Got shutdown signal, shutting down webhook server ...")
	server.Shutdown(context.Background())
}
