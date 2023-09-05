package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"segmentify/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateSegment(slug string) (string, error) {
	const op = "storage.postgres.CreateSegment"

	stmt, err := s.db.Prepare("INSERT INTO segments(slug) VALUES($1) RETURNING slug")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var dbSlug string

	err = stmt.QueryRow(slug).Scan(&dbSlug)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentExists)
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return dbSlug, nil
}

func (s *Storage) GetSegment(slug string) (string, error) {
	const op = "storage.postgres.GetSegment"

	stmt, err := s.db.Prepare("SELECT slug FROM segments WHERE slug = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var dbSlug string

	err = stmt.QueryRow(slug).Scan(&dbSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentNotFound)
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return dbSlug, nil
}

func (s *Storage) DeleteSegment(slug string) error {
	const op = "storage.postgres.DeleteSegment"

	stmt, err := s.db.Prepare("DELETE FROM segments WHERE slug = $1")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(slug)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected after delete: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSegmentNotFound)
	}

	return nil
}
