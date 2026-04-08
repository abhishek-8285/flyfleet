package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/flyfleet/fleet-core-service/internal/domains/auth/service"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    ctx := context.Background()
    dbURL := "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
    
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        fmt.Printf("DB Error: %v\n", err)
        os.Exit(1)
    }
    defer pool.Close()
    
    // Hash the password using Argon2 (same as auth service)
    password := "Fleet_password123"
    hash, err := service.HashPassword(password)
    if err != nil {
        fmt.Printf("Hash Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Generated hash: %s\n", hash)
    
    // Update all users
    emails := []string{
        "admin@flyfleet.com",
        "org_admin@delta.io",
        "fleet_manager@delta.io",
        "dispatcher@delta.io",
        "operator@delta.io",
        "maintenance@delta.io",
        "analyst@delta.io",
        "support_agent@delta.io",
        "billing@delta.io",
        "driver@delta.io",
    }
    
    for _, email := range emails {
        _, err := pool.Exec(ctx, "UPDATE users SET password_hash = $1 WHERE email = $2", hash, email)
        if err != nil {
            fmt.Printf("Update error for %s: %v\n", email, err)
        } else {
            fmt.Printf("Updated: %s\n", email)
        }
    }
    
    fmt.Println("\nAll users updated successfully!")
}
