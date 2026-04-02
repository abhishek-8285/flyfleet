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
	db, err := sql.Open("postgres", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	orgID := "00000000-0000-0000-0000-000000000001"

	// Argon2 Admin
	createAdmin(ctx, db, "admin@flyfleet.com", "password", orgID, true)
	// Bcrypt Admin
	createAdmin(ctx, db, "admin_bcrypt@flyfleet.com", "password", orgID, false)
}

func createAdmin(ctx context.Context, db *sql.DB, email, password, orgID string, useArgon2 bool) {
	var hash string
	if useArgon2 {
		salt := make([]byte, Argon2SaltLength)
		rand.Read(salt)
		h := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))
		b64Salt := base64.RawStdEncoding.EncodeToString(salt)
		b64Hash := base64.RawStdEncoding.EncodeToString(h)
		hash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)
	} else {
		// Manual static Bcrypt hash for "password"
		hash = "$2a$10$8K1p/90Wv.39WAXshq6fO.eTo9Ozyu8SfSjyv9V.SshWunpkZsj0."
	}

	// Clean up
	db.ExecContext(ctx, "DELETE FROM user_organizations WHERE user_id IN (SELECT id FROM users WHERE email = $1)", email)
	db.ExecContext(ctx, "DELETE FROM users WHERE email = $1", email)

	userID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, email, hash, "System Admin", "active", time.Now(), time.Now())
	
	if err != nil {
		log.Printf("Failed to insert user %s: %v", email, err)
		return
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO user_organizations (user_id, org_id, role, status)
		VALUES ($1, $2, $3, $4)`,
		userID, orgID, "admin", "active")

	if err != nil {
		log.Printf("Failed to link org for %s: %v", email, err)
		return
	}

	fmt.Printf("Created user %s (Argon2=%v)\n", email, useArgon2)
}
