package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Global database connection pool
var DB *pgxpool.Pool

// ConnectDB connects the Go backend to PostgreSQL
func ConnectDB() error {

	// 1. Load environment variables from .env
	if err := godotenv.Load("../../.env"); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// 2. Read database configuration
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// 3. Check required environment variables
	if host == "" || port == "" || user == "" ||
		password == "" || dbname == "" || sslmode == "" {
		return fmt.Errorf("missing database environment variables")
	}

	// 4. Create connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)

	// 5. Set connection timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// 6. Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	// 7. Test database connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 8. Save connection globally
	DB = pool

	fmt.Println("Connected to PostgreSQL successfully!")

	return nil
}
