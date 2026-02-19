package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("your-jwt-secret-key-here") // fallback for development only
		fmt.Println("WARNING: Using default secret. Set JWT_SECRET environment variable for production.")
	}

	claims := jwt.MapClaims{
		"user_id": "00000000-0000-0000-0000-000000000001",
		"org_id":  "00000000-0000-0000-0000-000000000001",
		"email":   "admin@flyfleet.io",
		"role":    "super_admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println(tokenString)
}
