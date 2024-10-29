package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/cmd/api"
	"github.com/sikozonpc/ecom/db"
	"github.com/stripe/stripe-go/v72"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}


	cfg := db.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
        SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	conn, err := db.NewPostgresStorage(cfg)

	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("Could not ping PostgreSQL: %v", err)
	}

	log.Println("Connected to PostgreSQL Database!")

	api_key := os.Getenv("STRIPE_SECRET_KEY")
	if api_key == "" {
		log.Fatal("Missing SECRET_KEY environment variable")
	}

	stripe.Key = api_key

	server := api.NewAPIServer(
		":8000",
		conn,
	)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}