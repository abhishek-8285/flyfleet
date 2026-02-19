package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("super-secret-key") // fallback for development only
		fmt.Println("WARNING: Using default secret. Set JWT_SECRET environment variable for production.")
	}
	
	orgID, _ := uuid.Parse("1476a987-76ad-4963-b64a-01a71eb739ca")
	userID := uuid.New()

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"org_id":  orgID.String(),
		"role":    "super_admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("%s|%s", tokenString, orgID.String())
}
