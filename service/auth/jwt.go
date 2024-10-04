package auth

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/db"
)

type contextKey string

const UserKey contextKey = "user_id"

func CreateJWT(secret []byte, userID int) (string, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	ext := db.PostgresConfig{
		JWTExpiration: db.ParseJWTExp(os.Getenv("JWT_EXPIRATION")),
	}

	expiration := time.Second * time.Duration(ext.JWTExpiration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": strconv.Itoa(userID),
        "exp":    time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}