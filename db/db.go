package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Event is a flattened view of an event row joined with its properties.
type Event struct {
	ID        string     `json:"id"`
	Event     string     `json:"event"`
	UserID    string     `json:"user_id"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Path      *string    `json:"path,omitempty"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	Total     *float64   `json:"total,omitempty"`
	Page      *string    `json:"page,omitempty"`
}

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

func PruneDB() error {
	if DB == nil {
		return errors.New("DB not initialised")
	}

	query := `DELETE FROM events;`
	_, err := DB.Exec(query)

	return err
}

// QueryEvents returns the most recent `limit` events joined with their properties.
func QueryEvents(limit int) ([]Event, error) {
	if DB == nil {
		return nil, errors.New("DB not initialised")
	}

	rows, err := DB.Query(`
		SELECT e.id, e.event, e.user_id, e.timestamp, e.path, e.type, e.created_at,
		       p.total, p.page
		FROM events e
		LEFT JOIN properties p ON e.id = p.event_id
		ORDER BY e.created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var timestamp sql.NullTime
		var path sql.NullString
		var total sql.NullFloat64
		var page sql.NullString

		if err := rows.Scan(&e.ID, &e.Event, &e.UserID, &timestamp, &path, &e.Type, &e.CreatedAt, &total, &page); err != nil {
			return nil, err
		}

		if timestamp.Valid {
			e.Timestamp = &timestamp.Time
		}
		if path.Valid {
			e.Path = &path.String
		}
		if total.Valid {
			e.Total = &total.Float64
		}
		if page.Valid {
			e.Page = &page.String
		}

		events = append(events, e)
	}

	return events, nil
}