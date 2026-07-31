package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func main() {

		// We'll use a known ID if possible, but for now we just try to list them
	url := "http://localhost:3000/api/v1/admin/vehicles"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJmbHlmbGVldC1hcGkiLCJkZXZpY2VfZmluZ2VycHJpbnQiOiIzOWY5MTg3OGY3NDg0MmRmMzkxYWVhNmEzZGM1MjU4NThhNmI5ZjZhNGYxODMzYmNjOGQ0ZmI0Njg1NGRlODZiIiwiZXhwIjoxNzc2OTQwMTE4LCJpYXQiOjE3NzY5MzkyMTgsImlzcyI6ImZseWZsZWV0LWF1dGgtc2VydmljZSIsImp0aSI6ImNkMDczYWY1LWJjMTctNDI4Yi1hMmZmLTY3YWNjZTA4NDEzMSIsIm5iZiI6MTc3NjkzOTIxOCwicGVybWlzc2lvbnNfbWFzayI6OTIyMzM3MjAzNjg1NDc3NTgwNywicm9sZSI6Im9yZ19hZG1pbiIsInN1YiI6IjVkNmQzYTQ2LTI5ZTQtNDIzNS1hNDBkLTAzZTA4N2VjZDhiNSIsInRlbmFudF9pZCI6IjU1MGU4NDAwLWUyOWItNDFkNC1hNzE2LTQ0NjY1NTQ0MDAwMiIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VyX2lkIjoiNWQ2ZDNhNDYtMjllNC00MjM1LWE0MGQtMDNlMDg3ZWNkOGI1In0.O9l0UNVfxcZ6v826CSGWgChSqeHErb0njV_IKjCrLkE")
    // Using the token from the user request history if possible, or a dummy one.
    // Actually I can't easily get a valid token here.
    
    // I'll just check the DB directly for the last updated vehicle.
	conn, err := pgx.Connect(context.Background(), "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM telemetry_points WHERE device_id = $1", "1234567890").Scan(&count)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	fmt.Printf("Simulated rows (Device: 1234567890) in telemetry_points: %d\n", count)
}
