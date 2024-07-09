package rpcconn

import (
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
)

type TransportType string

const ServerTransportType TransportType = "SERVER"
const ClientTransportType TransportType = "CLIENT"

func NewTransportCredential(transportType TransportType) (*credentials.TransportCredentials, error) {
	caCertFile := os.Getenv("CA_CERT_FILE")
	mTlsEnabled := os.Getenv("TLS_ENABLE_MTLS") == "true"
	tlsSkipVerify := os.Getenv("TLS_INSECURE_SKIP_VERIFY") == "true"
	transportTypeString := string(transportType)
	certFile := os.Getenv(transportTypeString + "_CERT_FILE")
	keyFile := os.Getenv(transportTypeString + "_KEY_FILE")

	if certFile != "" || keyFile != "" {
		log.Println("TLS enabled", certFile, keyFile)
		keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keyPair},
		}

		if caCertFile != "" {
			log.Println("Using mTLS with certificate authority", caCertFile)
			caCertData, err := os.ReadFile(caCertFile)
			if err != nil {
				return nil, err
			}

			caPool := x509.NewCertPool()
			ok := caPool.AppendCertsFromPEM(caCertData)
			if !ok {
				return nil, err
			}

			tlsConfig.ClientCAs = caPool
			tlsConfig.RootCAs = caPool
		}

		if mTlsEnabled {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		if tlsSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}

		transportCredentials := credentials.NewTLS(&tlsConfig)
		return &transportCredentials, nil
	} else {
		return nil, nil
	}
}
