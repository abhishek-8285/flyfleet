package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	email := "admin@delta.io"
	var userID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		log.Fatalf("User not found: %v", err)
	}
	fmt.Printf("User ID: %s\n", userID)

	// Check user_tenants
	rows, err := pool.Query(ctx, "SELECT tenant_id, role, role_id, is_active FROM user_tenants WHERE user_id = $1", userID)
	if err != nil {
		log.Fatalf("Error querying user_tenants: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nUser Tenants:")
	for rows.Next() {
		var tenantID uuid.UUID
		var role string
		var roleID *uuid.UUID
		var isActive bool
		if err := rows.Scan(&tenantID, &role, &roleID, &isActive); err != nil {
			log.Fatalf("Error scanning user_tenants: %v", err)
		}
		fmt.Printf("Tenant: %s, Role: %s, RoleID: %v, Active: %v\n", tenantID, role, roleID, isActive)
	}

	// Check user_roles
	rows, err = pool.Query(ctx, "SELECT tenant_id, role_id FROM user_roles WHERE user_id = $1", userID)
	if err != nil {
		log.Fatalf("Error querying user_roles: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nUser Roles:")
	for rows.Next() {
		var tenantID uuid.UUID
		var roleID uuid.UUID
		if err := rows.Scan(&tenantID, &roleID); err != nil {
			log.Fatalf("Error scanning user_roles: %v", err)
		}
		fmt.Printf("Tenant: %s, RoleID: %s\n", tenantID, roleID)

		// Check role details
		var roleName string
		var mask int64
		err = pool.QueryRow(ctx, "SELECT name, permission_mask FROM roles WHERE id = $1", roleID).Scan(&roleName, &mask)
		if err == nil {
			fmt.Printf("  Role Name: %s, Mask: %d\n", roleName, mask)
		}
	}
}
