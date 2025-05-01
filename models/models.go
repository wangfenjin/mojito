// Package models contains the data models and database access layer
package models

import (
	"context"
	"fmt"
	"sync" // Import sync package for thread safety

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wangfenjin/mojito/models/gen"
)

// globalDB holds the global database connection instance
var (
	globalDB *DB
	once     sync.Once // Use sync.Once to ensure Connect is called only once for initialization
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

// Connect establishes a connection to the database and initializes the global instance
func Connect(params ConnectionParams) (*DB, error) {
	var err error
	once.Do(func() { // Ensure this block runs only once
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
			params.User, params.Password, params.Host, params.Port, params.DBName, params.SSLMode, params.TimeZone,
		)

		conn, connErr := pgx.Connect(context.Background(), dsn)
		if connErr != nil {
			err = fmt.Errorf("failed to connect to PostgreSQL database: %w", connErr)
			return
		}

		// Create queries with the database connection
		queries := gen.New(conn)

		// Assign the created DB instance to the global variable
		globalDB = &DB{
			Conn:    conn,
			Queries: queries,
		}
	})
	return globalDB, err
}

// GetDB returns the globally initialized database instance.
// It panics if the database is not initialized.
func GetDB() *DB {
	if globalDB == nil {
		// Or return an error: return nil, fmt.Errorf("database not initialized")
		panic("database connection has not been initialized. Call Connect first.")
	}
	return globalDB
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(ctx context.Context, fn func(*gen.Queries) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := db.Queries.WithTx(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
