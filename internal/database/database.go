package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a connection pool to the database at the given URL and verifies it is rachable with a ping

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error){
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("crating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("Pinging database: %w", err)
	}

	return pool, nil 
}