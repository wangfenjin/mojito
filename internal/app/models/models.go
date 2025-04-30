// Package models contains the data models and database access layer
package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wangfenjin/mojito/internal/app/models/gen"
)

// ConnectionParams holds the parameters for connecting to the database
type ConnectionParams struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// DB wraps the database connection and queries
type DB struct {
	*pgx.Conn
	*gen.Queries
}

// Connect establishes a connection to the database
func Connect(params ConnectionParams) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
		params.User, params.Password, params.Host, params.Port, params.DBName, params.SSLMode, params.TimeZone,
	)

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	// Create queries with the database connection
	queries := gen.New(conn)

	return &DB{
		Conn:    conn,
		Queries: queries,
	}, nil
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(ctx context.Context, fn func(*gen.Queries) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := gen.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
