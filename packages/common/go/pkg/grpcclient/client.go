package grpcclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/fleet-platform/shared/pkg/resilience"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a gRPC client wrapper
type Client struct {
	Conn    *grpc.ClientConn
	Policy  *resilience.Policy
	Timeout time.Duration
}

// NewClient creates a new gRPC client with resilience policy.
// It will attempt to use mTLS if TLS_CERT_PATH, TLS_KEY_PATH, and TLS_CA_PATH are configured in the environment.
// Otherwise, it will fall back to insecure transport.
func NewClient(target string, timeout time.Duration) (*Client, error) {
	const maxMessageSize = 4 * 1024 * 1024
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	}

	creds, err := loadTLSCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials for client %s: %w", target, err)
	}

	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	policy := resilience.NewPolicy(resilience.DefaultConfig(target))

	return &Client{
		Conn:    conn,
		Policy:  policy,
		Timeout: timeout,
	}, nil
}

// Call executes a gRPC call with resilience logic
func (c *Client) Call(ctx context.Context, action func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	return c.Policy.Execute(ctx, func() error {
		return action(ctx)
	})
}

// Close closes the connection
func (c *Client) Close() error {
	return c.Conn.Close()
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	caPath := os.Getenv("TLS_CA_PATH")
	certPath := os.Getenv("TLS_CERT_PATH")
	keyPath := os.Getenv("TLS_KEY_PATH")

	if caPath == "" || certPath == "" || keyPath == "" {
		return nil, nil // Return nil, caller should use insecure
	}

	// 1. Load client cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client TLS certificate: %w", err)
	}

	// 2. Load CA to verify the server
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load root CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append root CA to certificate pool")
	}

	// 3. Create TLS config requiring client auth verified against CA pool
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS13,
		// InsecureSkipVerify is purposefully missing to enforce strict Server Name Indication (SNI) checks against SANs.
	}

	return credentials.NewTLS(tlsConfig), nil
}
