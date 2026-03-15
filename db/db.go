package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

type PgConfig struct {
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_DB       string
	POSTGRES_HOST     string
	POSTGRES_PORT     string
}

func loadEnv() *PgConfig {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env")
	}

	return &PgConfig{
		POSTGRES_USER:     strings.TrimSpace(os.Getenv("POSTGRES_USER")),
		POSTGRES_PASSWORD: strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")),
		POSTGRES_DB:       strings.TrimSpace(os.Getenv("POSTGRES_DB")),
		POSTGRES_HOST:     strings.TrimSpace(os.Getenv("POSTGRES_HOST")),
		POSTGRES_PORT:     strings.TrimSpace(os.Getenv("POSTGRES_PORT")),
	}
}

func (c *PgConfig) toConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.POSTGRES_HOST, c.POSTGRES_PORT, c.POSTGRES_USER, c.POSTGRES_PASSWORD, c.POSTGRES_DB)
}

func ConnectDB() {
	config := loadEnv()

	db, err := sql.Open("postgres", config.toConnectionString())

	if err != nil {
		panic(fmt.Sprintf("failed to open db connection: %s", err))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to reach postgres: %s", err))
	}

	DB = db
}
