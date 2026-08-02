package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/flyfleet/fleet-core/internal/domains/iam"
	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "ERROR: JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	// Initialize IAM Module
	iamModule := iam.NewModule(pool, nil, jwtSecret, config.DefaultIAMConfig(), logger)

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
	loginReq := models.LoginRequest{
		Email:    email,
		Password: password,
	}

	tokens, mfaResp, err := iamModule.AuthService.Login(ctx, loginReq, "127.0.0.1", "diagnostic-tool")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)

		user, _ := iamModule.UserRepo.GetByEmail(ctx, email)
		if user != nil {
			// SECURITY FIX: Do not print password hashes to console
			fmt.Printf("User found in DB.\n")
		} else {
			fmt.Println("User NOT found in DB")
		}
		return
	}

	if tokens != nil {
		fmt.Printf("Login successful! User ID: %s, Role: %s\n", tokens.User.ID, tokens.User.Role)
		if tokens.AccessToken != "" {
			fmt.Printf("ACCESS_TOKEN=%s\n", tokens.AccessToken)
		}
	} else if mfaResp != nil && mfaResp.TempToken != "" {
		fmt.Printf("MFA required. TEMP_TOKEN=%s\n", mfaResp.TempToken)
	}
}
