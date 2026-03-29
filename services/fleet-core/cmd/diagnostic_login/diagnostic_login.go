package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// SECURITY: Database connection must be explicitly configured
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "ERROR: DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Initialize Logger
	logConfig := logger.DefaultConfig()
	logConfig.Development = true
	l, err := logger.New(logConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Repositories and Services
	repo := repository.NewAuthRepository(pool, l)

	// IdentityService requires a *zap.Logger for its implementation (IdentityServiceImpl)
	// But it takes AuthRepositoryInterface
	// Let's create it.
	// We need to get the underlying *zap.Logger from ZapLogger if we can, or just mock it.
	// ZapLogger in packages/common/go has a private 'logger' field.
	// However, we can use the same logic here or just a dummy zap logger.

	identitySvc := service.NewIdentityService(repo, nil) // passing nil logger to see if it works

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "ERROR: JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	authSvc := service.NewAuthService(l, repo, identitySvc, jwtSecret, nil, nil)

	// SECURITY: Diagnostic login requires explicit credentials via environment
	// This tool is for legitimate debugging/administration only.
	// DO NOT commit credentials to source control.
	email := os.Getenv("DIAGNOSTIC_USER_EMAIL")
	if email == "" {
		fmt.Fprintln(os.Stderr, "ERROR: DIAGNOSTIC_USER_EMAIL environment variable is required")
		os.Exit(1)
	}

	password := os.Getenv("DIAGNOSTIC_USER_PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "ERROR: DIAGNOSTIC_USER_PASSWORD environment variable is required")
		os.Exit(1)
	}

	fmt.Printf("Attempting diagnostic login for %s...\n", email)
	resp, err := authSvc.Login(ctx, email, password, "", "127.0.0.1", "test-agent")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)

		user, _ := repo.GetUserByEmail(ctx, email)
		if user != nil {
			// SECURITY FIX: Do not print password hashes to console
			fmt.Printf("User found.\n")
		} else {
			fmt.Println("User NOT found in DB")
		}
		return
	}

	fmt.Printf("Login successful! User ID: %s, Role: %s, MFA Required: %v\n", resp.User.ID, resp.User.Role, resp.MfaRequired)
	if resp.AccessToken != "" {
		fmt.Printf("ACCESS_TOKEN=%s\n", resp.AccessToken)
	} else if resp.TempToken != "" {
		fmt.Printf("TEMP_TOKEN=%s (MFA required)\n", resp.TempToken)
	}
}
