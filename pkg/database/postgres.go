package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

type DB struct {
	*sql.DB
}

func NewPostgresConnection(url string, maxOpenConns, maxIdleConns int) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("Database connection established")

	return &DB{db}, nil
}

func (db *DB) Close() error {
	logger.Info().Msg("Closing database connection")
	return db.DB.Close()
}

func (db *DB) HealthCheck() error {
	return db.Ping()
}
