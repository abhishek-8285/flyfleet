package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"github.com/joho/godotenv"
)

func main() {
	// Automatically load .env if it exists in the current directory
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is not set. Please ensure it is loaded in your shell.")
	}

	backupFile := "flyfleet_full_sync.sql"
	fmt.Printf("🚀 Starting High-Fidelity SQL Backup to %s...\n", backupFile)

	// pg_dump flags explained:
	// --no-owner: Removes ownership commands (essential for importing on different devices/users)
	// --no-acl: Removes access control lists/privileges
	// -f: Specifies the output file
	cmd := exec.Command("pg_dump", "--no-owner", "--no-acl", "-f", backupFile, dbURL)

	// Stream stderr to our console so we can see any issues
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Backup failed! Ensure 'pg_dump' is installed on this device: %v", err)
	}

	fi, err := os.Stat(backupFile)
	if err != nil {
		log.Fatalf("❌ Failed to verify backup file: %v", err)
	}

	fmt.Printf("\n✅ Full SQL Backup Complete!\n")
	fmt.Printf("📁 File: %s\n", backupFile)
	fmt.Printf("💾 Size: %.2f MB\n", float64(fi.Size())/(1024*1024))
	fmt.Printf("🚀 DONE: You can now copy this file to any other device and import it using: psql -d [new_db] -f %s\n", backupFile)
}
