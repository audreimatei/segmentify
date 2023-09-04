package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateSegment(slug string) (models.Segment, error) {
	const op = "storage.postgres.CreateSegment"
	defaultSegment := models.Segment{ID: 0, Slug: ""}

	stmt, err := s.db.Prepare("INSERT INTO segments(slug) VALUES($1) RETURNING id")
	if err != nil {
		return defaultSegment, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var resID int64

	err = stmt.QueryRow(slug).Scan(&resID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return defaultSegment, fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentExists)
		}

		return defaultSegment, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return models.Segment{ID: resID, Slug: slug}, nil
}

func (s *Storage) GetSegment(slug string) (models.Segment, error) {
	const op = "storage.postgres.GetSegment"
	defaultSegment := models.Segment{ID: 0, Slug: ""}

	stmt, err := s.db.Prepare("SELECT id FROM segments WHERE slug = $1")
	if err != nil {
		return defaultSegment, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var resID int64

	err = stmt.QueryRow(slug).Scan(&resID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultSegment, fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentNotFound)
		}

		return defaultSegment, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return models.Segment{ID: resID, Slug: slug}, nil
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
