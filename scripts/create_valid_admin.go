package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/argon2"
	"github.com/google/uuid"
)

const (
	Argon2Iterations  = 3
	Argon2Memory      = 64 * 1024
	Argon2Parallelism = 4
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

func main() {
	password := "password"
	email := "admin@flyfleet.com"
	orgID := "00000000-0000-0000-0000-000000000001"

	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		log.Fatal(err)
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))

	// RawStdEncoding is key here
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)

	db, err := sql.Open("postgres", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Disable RLS for user setup
	_, _ = db.ExecContext(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY")

	// 2. Clean up existing
	_, _ = db.ExecContext(ctx, "DELETE FROM user_organizations WHERE user_id IN (SELECT id FROM users WHERE email = $1)", email)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE email = $1", email)

	// 3. Insert fresh user (Matching exact schema)
	userID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, email, encoded, "System Admin", "active", time.Now(), time.Now())
	
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}

	// 4. Link to Org
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_organizations (user_id, org_id, role, status)
		VALUES ($1, $2, $3, $4)`,
		userID, orgID, "admin", "active")

	if err != nil {
		log.Fatalf("Failed to link org: %v", err)
	}

	// 5. Re-enable RLS
	_, _ = db.ExecContext(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")

	fmt.Printf("Successfully created valid admin user %s with Argon2id hash (v=%d).\n", email, argon2.Version)
}
