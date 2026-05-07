#!/bin/bash
export DATABASE_URL=postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable
go run cmd/migrate/main.go up
