//go:build linux
// +build linux

package server

import (
	"context"
	"crypto/tls"
	"github.com/g-portal/metadata-server/pkg/config"
	"github.com/g-portal/metadata-server/pkg/helper"
	"github.com/g-portal/metadata-server/pkg/logging"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

var server *http.Server

// StartServer Starts the Gateway server
func StartServer() {
	c := config.GetConfig()
	log.Printf("Starting Latency Service on %s.", c.Hostname)

	if c.Logging {
		log.Printf("Enabled logging to %s.", c.GetLogFile())
		logging.SetLogFile(path.Join(c.GetLogFile()))
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if c.LetsEncrypt {
		if !helper.FileExists(c.GetCertificateDirectory()) {
			err := os.Mkdir(c.GetCertificateDirectory(), 0700)
			if err != nil {
				log.Fatalln(err)
			}
		}

		certManager := autocert.Manager{
			Prompt:      autocert.AcceptTOS,
			Email:       c.LetsEncryptEmail,
			HostPolicy:  autocert.HostWhitelist(c.Hostname),
			Cache:       autocert.DirCache(c.GetCertificateDirectory()),
			RenewBefore: 12 * time.Hour,
			Client: &acme.Client{
				DirectoryURL: acme.LetsEncryptURL,
			},
		}

		tlsConfig.GetCertificate = certManager.GetCertificate

		// Create a listener
		log.Printf("Starting HTTP service on %s", c.ListenAddressHTTP)
		httpListener, err := net.Listen("tcp", c.ListenAddressHTTP)
		if err != nil {
			log.Fatalln(err)
		}

		// Start HTTP server only for ACME http-01 auth
		go func() {
			err := http.Serve(httpListener, certManager.HTTPHandler(nil))
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}

	log.Printf("Starting HTTPS service on %s", c.ListenAddressHTTPS)
	httpsListener, err := net.Listen("tcp", c.ListenAddressHTTPS)
	if err != nil {
		log.Fatalln(err)
	}

	server = &http.Server{
		TLSConfig:   tlsConfig,
		ConnContext: saveConnIntoContext,
	}

	// Register gateway handler
	http.Handle("/ping", ping())

	// Start HTTPS server
	if c.LetsEncrypt {
		err = server.ServeTLS(httpsListener, "", "")
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		err = server.Serve(httpsListener)
		if err != nil {
			log.Fatalln(err)
		}
	}

}

// StopServer Stops the HTTP server
func StopServer() {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
