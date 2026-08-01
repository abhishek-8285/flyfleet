package main

import (
	"fmt"
	"os"

	"github.com/gonotify/gonotify/internal/auth"
)

func main() {
	environment := "production"

	if len(os.Args) > 1 {
		environment = os.Args[1]
	}

	raw, prefix, hash, err := auth.GenerateAPIKey(environment)
	if err != nil {
		panic(err)
	}

	fmt.Printf("RAW_API_KEY=%s\n", raw)
	fmt.Printf("KEY_PREFIX=%s\n", prefix)
	fmt.Printf("KEY_HASH=%s\n", hash)
}
