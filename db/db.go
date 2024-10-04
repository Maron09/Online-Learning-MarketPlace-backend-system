package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	Host          string
	Port          string
	User          string
	Password      string
	DBName        string
	SSLMode       string
	JWTExpiration int64
	JWTSecret     string
}

func NewPostgresStorage(cfg PostgresConfig) (*sql.DB, error) {
	psqlConStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", psqlConStr)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}


func ParseJWTExp(expStr string) int64 {
	expiration, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid JWT expiration value: %v", err)
	}
	return expiration
}