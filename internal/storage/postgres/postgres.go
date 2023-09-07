package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	fail := func(msg string, err error) (*Storage, error) {
		return nil, fmt.Errorf("storage.postgres.New: %s: %w", msg, err)
	}

	pool, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return fail("failed to create a database poll", err)
	}

	connAttempts := 10

	for connAttempts > 0 {
		time.Sleep(time.Second)

		err = pool.Ping(ctx)
		if err == nil {
			break
		}

		connAttempts--
	}

	if err != nil {
		return fail("failed to ping a database", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	const op = "storage.postgres.Init"

	f, err := os.Open("internal/storage/postgres/init.sql")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer f.Close()

	query, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.pool.Exec(ctx, string(query))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
