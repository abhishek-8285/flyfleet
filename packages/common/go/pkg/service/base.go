package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/fleet-platform/shared/pkg/telemetry"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// Base represents the standard foundation for all Go microservices.
type Base struct {
	Name    string
	Logger  logger.Logger
	DB      database.Database
	Config  *viper.Viper
	cleanup func()
}

// Config represents standard connection parameters
type Config struct {
	DatabaseURL string
	LogLevel    string
	Port        string
	GRPCPort    string
	TLSCACert    string
	TLSServer    string
	TLSKey       string
	OTLPEndpoint string
}

// NewBase initializes the standard service template: logger, config, and DB pooling.
func NewBase(serviceName string, defaultCfg Config) (*Base, error) {
	// 1. Setup Configuration via Viper
	v := viper.New()
	v.SetDefault("server.port", defaultCfg.Port)
	v.SetDefault("server.grpc_port", defaultCfg.GRPCPort)
	v.SetDefault("server.log_level", defaultCfg.LogLevel)
	v.SetDefault("database.url", defaultCfg.DatabaseURL)
	v.SetDefault("tls.ca", defaultCfg.TLSCACert)
	v.SetDefault("tls.cert", defaultCfg.TLSServer)
	v.SetDefault("tls.key", defaultCfg.TLSKey)
	v.SetDefault("telemetry.otlp_endpoint", defaultCfg.OTLPEndpoint)

	// In Kubernetes and Docker we mount secrets to /run/secrets
	v.AddConfigPath("/run/secrets")
	v.AddConfigPath(".")
	
	// Read standard config if available (optional)
	_ = v.ReadInConfig()

	v.AutomaticEnv()
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.grpc_port", "GRPC_PORT")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("tls.ca", "TLS_CA_PATH")
	v.BindEnv("tls.cert", "TLS_CERT_PATH")
	v.BindEnv("tls.key", "TLS_KEY_PATH")
	v.BindEnv("telemetry.otlp_endpoint", "OTLP_ENDPOINT")

	// Map Individual Secrets (e.g. database_url as a file in /run/secrets/database_url)
	secretKeys := []string{
		"DATABASE_URL",
		"DB_PASSWORD",
		"JWT_SECRET",
		"API_KEY",
	}
	for _, key := range secretKeys {
		secretPath := "/run/secrets/" + key
		if val, err := os.ReadFile(secretPath); err == nil {
			v.Set(key, string(bytes.TrimSpace(val))) // Allow code to read raw secret values
		}
	}

	// 2. Setup Logger
	logCfg := logger.DefaultConfig()
	logCfg.Development = v.GetString("server.log_level") == "debug"

	sysLog, err := logger.New(logCfg)
	if err != nil {
		return nil, err
	}

	// 3. Setup Database with shared pgx wrapper
	var db database.Database
	dbURL := v.GetString("database.url")

	if dbURL != "" {
		db, err = database.NewPostgres(dbURL, sysLog.(*logger.ZapLogger).WithContext(context.Background()))
		if err != nil {
			sysLog.Error("Failed to initialize database pool", zap.Error(err))
			return nil, err
		}
	}

	// 4. Setup Telemetry
	shutdownOTel, err := telemetry.InitProvider(context.Background(), sysLog, serviceName, v.GetString("telemetry.otlp_endpoint"))
	if err != nil {
		sysLog.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		// Optional: We can continue without telemetry or fail. We'll fail here to be strict since OTLP_ENDPOINT was set.
		if v.GetString("telemetry.otlp_endpoint") != "" {
			return nil, err
		}
	}

	return &Base{
		Name:   serviceName,
		Logger: sysLog,
		DB:     db,
		Config: v,
		cleanup: func() {
			sysLog.Info("Shutting down service resources")
			if db != nil {
				_ = db.Close()
			}
			if shutdownOTel != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := shutdownOTel(ctx); err != nil {
					sysLog.Error("Failed to completely shutdown OpenTelemetry", zap.Error(err))
				}
			}
		},
	}, nil
}

// GetTLSCredentials constructs the TransportCredentials to safely host gRPC over mTLS
func (b *Base) GetTLSCredentials() (credentials.TransportCredentials, error) {
	caPath := b.Config.GetString("tls.ca")
	certPath := b.Config.GetString("tls.cert")
	keyPath := b.Config.GetString("tls.key")

	if caPath == "" || certPath == "" || keyPath == "" {
		b.Logger.Warn("TLS configuration incomplete, falling back to insecure transport in development mode")
		return nil, nil // Return nil, up to implementer to use insecure.NewCredentials
	}

	// 1. Load Server Certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server TLS certificate: %w", err)
	}

	// 2. Load Root CA (used to verify the client requesting access)
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
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS13, // Enforce modern crypto only
	}

	return credentials.NewTLS(tlsConfig), nil
}

// ServeHealth starts a standardized background HTTP server for liveness and readiness probes
func (b *Base) ServeHealth(port string) *http.Server {
	mux := http.NewServeMux()

	// Liveness probe (is process running?)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness probe (can process traffic?)
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if b.DB != nil {
			// A simple lightweight query to check connection health instead of ping
			_, err := b.DB.Query(context.Background(), "SELECT 1")
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Database unavailable"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: telemetry.WrapHTTPHandler(mux, "health_probe"),
	}

	go func() {
		b.Logger.Info("Starting standard health probe server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.Logger.Error("Health server failed", zap.Error(err))
		}
	}()

	return srv
}

// RunWithContext handles OS signals for graceful termination
func (b *Base) RunWithContext(ctx context.Context, runners ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer b.cleanup()

	// Watch for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, len(runners))

	// Start all runner functions
	for _, runner := range runners {
		r := runner // capture loop variable
		go func() {
			errCh <- r(ctx)
		}()
	}

	// Wait for signal or error
	select {
	case <-sigCh:
		b.Logger.Info("Received termination signal, initiating graceful shutdown")
		cancel() // Signal all runners to stop
	case err := <-errCh:
		if err != nil {
			b.Logger.Error("Service runner fatal error", zap.Error(err))
			cancel()
			return err
		}
	}

	// Optional: wait a bit for resources to cleanly close
	time.Sleep(1 * time.Second)
	b.Logger.Info("Graceful shutdown complete")
	return nil
}
