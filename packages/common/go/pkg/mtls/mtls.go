// Package mtls provides mutual TLS configuration helpers for gRPC
// service-to-service communication within the FlyFleet platform.
//
// Usage (server):
//
//	tlsCfg, err := mtls.ServerConfig("/certs/server.crt", "/certs/server.key", "/certs/ca.crt")
//	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)))
//
// Usage (client):
//
//	tlsCfg, err := mtls.ClientConfig("/certs/client.crt", "/certs/client.key", "/certs/ca.crt")
//	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// ServerConfig returns a tls.Config suitable for a gRPC server with mutual TLS.
// - certFile: path to the server's certificate (PEM)
// - keyFile:  path to the server's private key (PEM)
// - caFile:   path to the CA certificate that signed client certs
func ServerConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: load server cert: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: read CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("mtls: failed to parse CA cert")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// ClientConfig returns a tls.Config suitable for a gRPC client with mutual TLS.
// - certFile: path to the client's certificate (PEM)
// - keyFile:  path to the client's private key (PEM)
// - caFile:   path to the CA certificate that signed server certs
func ClientConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: load client cert: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: read CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("mtls: failed to parse CA cert")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// InsecureConfig returns a no-op TLS config for development environments.
// This should NEVER be used in production.
func InsecureConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
	}
}
