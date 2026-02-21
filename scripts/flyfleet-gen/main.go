package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: flyfleet-gen <command> <service-name>")
		fmt.Println("Commands: service")
		os.Exit(1)
	}

	command := os.Args[1]
	serviceName := os.Args[2]

	switch command {
	case "service":
		createService(serviceName)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func createService(name string) {
	fmt.Printf("Generating service: %s\n", name)

	dirs := []string{
		"cmd/server",
		"internal/config",
		"internal/database",
		"internal/events",
		"internal/grpc",
		"internal/handler",
		"internal/middleware",
		"internal/models",
		"internal/repository",
		"internal/service",
		"pkg/common",
		"migrations",
		"proto",
	}

	for _, dir := range dirs {
		path := filepath.Join("services", name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", path, err)
			return
		}
	}

	// Generate basic go.mod
	goModContent := fmt.Sprintf("module github.com/flyfleet/%s\n\ngo 1.24.0\n", name)
	if err := os.WriteFile(filepath.Join("services", name, "go.mod"), []byte(goModContent), 0644); err != nil {
		fmt.Printf("Failed to write go.mod: %v\n", err)
	}

	fmt.Printf("Successfully generated service template in services/%s\n", name)
}
