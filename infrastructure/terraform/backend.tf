terraform {
  backend "pg" {
    # Connection string for PostgreSQL backend
    # You must provide this via -backend-config or environment variable
    # Example: conn_str = "postgresql://terraform:password@localhost:5432/terraform_state?sslmode=disable"
    conn_str = ""
  }
}
