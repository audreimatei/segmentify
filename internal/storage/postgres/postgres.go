package postgres

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open the database: %w", op, err)
	}

	connAttempts := 10

	for connAttempts > 0 {
		time.Sleep(time.Second)

		err = db.Ping()
		if err == nil {
			break
		}

		connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to the database: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Init() error {
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

	_, err = s.db.Exec(string(query))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	const op = "storage.postgres.Close"

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
