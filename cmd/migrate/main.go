package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
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
	if err := conn.Ping(); err!= nil {
        log.Fatalf("Could not ping PostgreSQL: %v", err)
    }

	log.Println("Connected to PostgreSQL successfully")

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err!= nil {
        log.Fatal(err)
    }

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[1]

	switch cmd{
	case "up":
		if err := m.Up(); err!= nil && err!= migrate.ErrNoChange {
            log.Fatal(err)
        }
		log.Println("Migrations up to date")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migration rolled back successfully")
	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}