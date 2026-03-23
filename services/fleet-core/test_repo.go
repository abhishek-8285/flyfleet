package main

import (
	"context"
	"fmt"
	"log"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DummyLogger implements the logger.Logger interface minimally
type DummyLogger struct {
	*zap.Logger
}

func (l *DummyLogger) With(fields ...zap.Field) *zap.Logger {
	return l.Logger.With(fields...)
}

func main() {
	dbUrl := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	zlog, _ := zap.NewDevelopment()
	_ = &DummyLogger{Logger: zlog}
	
	// We might need to cast or just let Go duck wrap it
	// Actually repository.NewAuthRepository expects logger.Logger which likely has Info, Error etc. which zap.Logger has.
	// But let's see what happens.
	
	// Instead of passing interface, we can just run the query manually!
	query := "SELECT id, org_id, email, CAST(role_id AS VARCHAR) as role, password_hash, created_at, created_by, updated_at, updated_by, mfa_enabled, mfa_methods, mfa_secret, last_password_change, failed_login_attempts, locked_until FROM users WHERE email = $1"
	
	var id, orgId, email string
	var role *string
	err = pool.QueryRow(ctx, query, "abhishek@test1.com").Scan(
		&id, &orgId, &email, &role, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	
	repo := repository.NewAuthRepository(pool, nil)
	user, err := repo.GetUserByEmail(ctx, "abhishek@test1.com")
	
	if err != nil {
		fmt.Printf("GOT ERROR FROM REPO: %T %v\n", err, err)
		return
	}

	fmt.Printf("GOT USER: %+v\n", user)
}
