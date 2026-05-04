package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	query := `
	CREATE TABLE IF NOT EXISTS documents (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		entity_type VARCHAR(50) NOT NULL,
		entity_id UUID NOT NULL,
		document_type VARCHAR(50) NOT NULL,
		file_name VARCHAR(255) NOT NULL,
		file_path TEXT NOT NULL,
		file_size_bytes BIGINT NOT NULL,
		mime_type VARCHAR(100) NOT NULL,
		expiry_date TIMESTAMP WITH TIME ZONE,
		status VARCHAR(20) DEFAULT 'PENDING',
		remarks TEXT,
		metadata JSONB DEFAULT '{}',
		uploaded_by UUID,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);
	`
	_, err = pool.Exec(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Documents table ensured successfully.")
}
