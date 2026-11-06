package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)


var DB *pgxpool.Pool


func ConnectDB() error {

	
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		return fmt.Errorf("missing DATABASE_URL")
	}

	
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	
	DB = pool

	fmt.Println("Connected to PostgreSQL successfully!")

	return nil
}
