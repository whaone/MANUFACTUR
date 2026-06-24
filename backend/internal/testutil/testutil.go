//go:build integration

package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"manufactpro/backend/internal/db"
)

var pool *pgxpool.Pool

// Pool returns the shared test pool (valid after Setup).
func Pool() *pgxpool.Pool { return pool }

// Setup connects to the test DB and returns a cleanup func.
// Reads DATABASE_URL from env; falls back to the dev DB.
func Setup(t *testing.T) func() {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://whaone17:manufactpro@127.0.0.1:5434/manufactpro?sslmode=disable"
	}
	var err error
	pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("testutil: connect DB: %v", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		t.Fatalf("testutil: ping DB: %v", err)
	}
	db.Pool = pool
	return func() { pool.Close() }
}
