package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

const (
	Argon2Iterations  = 3
	Argon2Memory      = 64 * 1024
	Argon2Parallelism = 4
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

func main() {
	ctx := context.Background()
	connStr := "postgresql://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
	
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	email := "admin@flyfleet.com"
	password := "admin123"
	
	// Role: super_admin (matches frontend expectation)
	roleID := uuid.MustParse("38857329-fa25-4536-aa8d-832587f97735")
	// Org: Default Organization (from previous check)
	orgID := uuid.MustParse("ce95cfe5-42a2-48b2-99d5-3f817c2f0247")
	
	hash, err := generateArgon2Hash(password)
	if err != nil {
		log.Fatal(err)
	}

	var userID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		userID = uuid.New()
	}

	query := `
		INSERT INTO users (
			id, org_id, email, password_hash, role, role_id, status, 
			permission_mask, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (email) DO UPDATE SET 
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			role_id = EXCLUDED.role_id,
			permission_mask = EXCLUDED.permission_mask,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`
	
	err = pool.QueryRow(ctx, query,
		userID, orgID, email, hash, "super_admin", roleID, "active",
		9223372036854775807, time.Now(), time.Now(),
	).Scan(&userID)
	
	if err != nil {
		log.Fatal(err)
	}

	// Insert into user_organizations
	queryOrg := `
		INSERT INTO user_organizations (user_id, org_id, role, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, org_id) DO NOTHING
	`
	_, err = pool.Exec(ctx, queryOrg, userID, orgID, "super_admin", "active")
	if err != nil {
		log.Fatal(err)
	}

	// Insert into user_roles
	queryRole := `
		INSERT INTO user_roles (user_id, role_id, organization_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id, organization_id) DO NOTHING
	`
	_, err = pool.Exec(ctx, queryRole, userID, roleID, orgID)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Successfully seeded user %s with password %s\n", email, password)
}

func generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash), nil
}
