package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters (must match application constants from auth_service.go)
	argon2Iterations  = 3
	argon2Memory      = 64 * 1024 // 64 MB in KiB (argon2 takes KiB, so 64*1024 = 65536 KiB)
	argon2Parallelism = 4
	argon2SaltLength  = 16
	argon2KeyLength   = 32
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Email      string
	Password   string
	Name       string
}

func main() {
	cfg := Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     getEnv("POSTGRES_USER", "fleet_user"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "fleet_password"),
		DBName:     getEnv("POSTGRES_DB", "fleet_db"),
		Email:      getEnv("ADMIN_EMAIL", "admin@flyfleet.com"),
		Password:   getEnv("ADMIN_PASSWORD", "Admin123!"),
		Name:       getEnv("ADMIN_NAME", "FlyFleet Admin"),
	}

	// Override with command-line args if provided
	if len(os.Args) > 1 {
		cfg.Email = os.Args[1]
	}
	if len(os.Args) > 2 {
		cfg.Password = os.Args[2]
	}
	if len(os.Args) > 3 {
		cfg.Name = os.Args[3]
	}

	fmt.Println("========================================")
	fmt.Println("FlyFleet Admin User Creation")
	fmt.Println("========================================")
	fmt.Printf("Email: %s\n", cfg.Email)
	fmt.Printf("Name:  %s\n", cfg.Name)
	fmt.Println()

	// Build connection string
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Connect to database
	fmt.Println("Connecting to database...")
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	fmt.Println("✅ Connected")

	// Create admin user
	if err := createAdmin(pool, cfg); err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start the service: docker-compose up -d fleet-core")
	fmt.Println("2. Login with the credentials you set")
	fmt.Println()
}

func createAdmin(pool *pgxpool.Pool, cfg Config) error {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Will be ignored if committed

	// 1. Get or create default organization
	var orgID uuid.UUID
	queryOrg := "SELECT id FROM organizations WHERE name = 'Default Organization' LIMIT 1"
	err = tx.QueryRow(ctx, queryOrg).Scan(&orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create the default organization
			orgID = uuid.New()
			insertOrg := "INSERT INTO organizations (id, name) VALUES ($1, $2)"
			_, err = tx.Exec(ctx, insertOrg, orgID, "Default Organization")
			if err != nil {
				return fmt.Errorf("failed to create organization: %w", err)
			}
			fmt.Println("   Created default organization")
		} else {
			return fmt.Errorf("failed to query organization: %w", err)
		}
	} else {
		fmt.Printf("   Using existing organization: %s\n", orgID)
	}

	// 2. Check if user exists
	var existingID uuid.UUID
	queryUser := "SELECT id FROM users WHERE email = $1"
	err = tx.QueryRow(ctx, queryUser, cfg.Email).Scan(&existingID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check existing user: %w", err)
		}
		// User does not exist, continue to create
	} else {
		// User exists, update role to super_admin
		updateRole := `
			UPDATE users SET
				org_id = $1,
				role = 'super_admin',
				status = 'active',
				permission_mask = $2,
				updated_at = $3,
				mfa_enabled = false,
				mfa_methods = '{}'
			WHERE email = $4
		`
		_, err = tx.Exec(ctx, updateRole, orgID, 9223372036854775807, time.Now(), cfg.Email)
		if err != nil {
			return fmt.Errorf("failed to update existing user to super_admin: %w", err)
		}
		fmt.Println("   Updated existing user to super_admin")
		// Also ensure user_organizations entry exists
		_ = ensureUserOrg(tx, ctx, existingID, orgID)
		return tx.Commit(ctx)
	}

	// 3. Generate Argon2id hash
	fmt.Println("   Generating password hash...")
	passwordHash, err := generateArgon2Hash(cfg.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Insert new admin user
	userID := uuid.New()
	now := time.Now()
	queryInsert := `
		INSERT INTO users (
			id, org_id, email, password_hash, role,
			status, is_deleted,
			created_at, updated_at, created_by, updated_by,
			mfa_enabled, mfa_methods, last_password_change,
			failed_login_attempts, locked_until,
			permission_mask
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9, $10, $11,
			$12, $13, $14,
			$15, $16,
			$17
		)
	`
	_, err = tx.Exec(ctx, queryInsert,
		userID,                    // id
		orgID,                    // org_id
		cfg.Email,                // email
		passwordHash,             // password_hash
		"super_admin",            // role (string)
		"active",                 // status
		false,                    // is_deleted
		now,                      // created_at
		now,                      // updated_at
		userID,                   // created_by (self)
		userID,                   // updated_by (self)
		false,                    // mfa_enabled
		"{}",                     // mfa_methods (empty array)
		now,                      // last_password_change
		0,                        // failed_login_attempts
		nil,                      // locked_until (NULL)
		9223372036854775807,      // permission_mask (all permissions)
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	fmt.Printf("   Created user ID: %s\n", userID)
	fmt.Printf("   Password hashed with Argon2id\n")

	return tx.Commit(ctx)
}

// generateArgon2Hash creates an Argon2id hash in the format: $argon2id$v=19$m=...,t=...,p=...$salt$hash
func generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(argon2Iterations), uint32(argon2Memory), uint8(argon2Parallelism), uint32(argon2KeyLength))

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func ensureUserOrg(tx pgx.Tx, ctx context.Context, userID, orgID uuid.UUID) error {
	query := `
		INSERT INTO user_organizations (user_id, org_id, role, status, created_at, updated_at)
		VALUES ($1, $2, 'admin', 'active', NOW(), NOW())
		ON CONFLICT (user_id, org_id) DO UPDATE SET 
			role = 'admin',
			status = 'active',
			updated_at = NOW()
	`
	_, err := tx.Exec(ctx, query, userID, orgID)
	return err
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
