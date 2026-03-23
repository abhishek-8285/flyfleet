package main

import (
	"context"
	"fmt"
	"log"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/fleet-platform/shared/pkg/logger"
)

func main() {
	ctx := context.Background()
	dbURL := "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
	
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

	authSvc := service.NewAuthService(l, repo, identitySvc, "your-jwt-secret-key-here", nil, nil)

	email := "super_admin@flyfleet.com"
	password := "fleet_password"

	fmt.Printf("Attempting diagnostic login for %s...\n", email)
	resp, err := authSvc.Login(ctx, email, password, "", "127.0.0.1", "test-agent")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		
		user, _ := repo.GetUserByEmail(ctx, email)
		if user != nil {
			fmt.Printf("User found. Hash in DB: %s\n", user.PasswordHash)
		} else {
			fmt.Println("User NOT found in DB")
		}
		return
	}

	fmt.Printf("Login successful! User ID: %s, Role: %s, MFA Required: %v\n", resp.User.ID, resp.User.Role, resp.MfaRequired)
}
