package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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
		POSTGRES_USER:     os.Getenv("POSTGRES_USER "),
		POSTGRES_PASSWORD: os.Getenv("POSTGRES_PASSWORD"),
		POSTGRES_DB:       os.Getenv("POSTGRES_DB"),
		POSTGRES_HOST:     os.Getenv("POSTGRES_HOST"),
		POSTGRES_PORT:     os.Getenv("POSTGRES_PORT"),
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
