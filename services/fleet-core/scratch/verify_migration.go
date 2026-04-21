package scratch

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'invoices' 
		AND column_name IN ('template_type', 'vehicle_no', 'eway_bill_no')
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Found columns in 'invoices':")
	for rows.Next() {
		var col string
		rows.Scan(&col)
		fmt.Printf("- %s\n", col)
	}
}
