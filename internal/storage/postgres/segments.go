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

func (s *Storage) CreateSegment(segment models.Segment) (models.Segment, error) {
	const op = "storage.postgres.CreateSegment"

	tx, err := s.db.Begin()
	if err != nil {
		return models.Segment{}, fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO segments(slug, percent)
		VALUES($1, $2)
	`, segment.Slug, segment.Percent); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return models.Segment{}, fmt.Errorf("%s: %w", op, storage.ErrSegmentExists)
		}
		return models.Segment{}, fmt.Errorf("%s: insert segment: %w", op, err)
	}

	if segment.Percent > 0 {
		var usersCount int64
		if err := tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&usersCount); err != nil {
			return models.Segment{}, fmt.Errorf("%s: count users: %w", op, err)
		}

		usersToAddCount := usersCount * segment.Percent / 100
		usersToAdd, err := s.GetRandomUsers(usersToAddCount)
		if err != nil {
			return models.Segment{}, fmt.Errorf("%s: get random users: %w", op, err)
		}

		stmt, err := tx.Prepare(`
			INSERT INTO users_segments(user_id, segment_slug, expire_at)
			VALUES($1, $2, $3)
		`)
		if err != nil {
			return models.Segment{}, fmt.Errorf("%s: prepare statement for add: %w", op, err)
		}
		defer stmt.Close()

		historyStmt, err := tx.Prepare(`
			INSERT INTO users_segments_history(user_id, segment_slug, operation)
			VALUES($1, $2, $3)
		`)
		if err != nil {
			return models.Segment{}, fmt.Errorf("%s: prepare statement for history: %w", op, err)
		}
		defer historyStmt.Close()

		for _, userID := range usersToAdd {
			if _, err = stmt.Exec(userID, segment.Slug, nil); err != nil {
				return models.Segment{}, fmt.Errorf("%s: insert user segment: %w", op, err)
			}

			if _, err = historyStmt.Exec(userID, segment.Slug, "add"); err != nil {
				return models.Segment{}, fmt.Errorf("%s: insert user segment history: %w", op, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return models.Segment{}, fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return segment, nil
}

func (s *Storage) GetSegment(slug string) (models.Segment, error) {
	const op = "storage.postgres.GetSegment"

	stmt, err := s.db.Prepare("SELECT percent FROM segments WHERE slug = $1")
	if err != nil {
		return models.Segment{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var dbPercent int64

	err = stmt.QueryRow(slug).Scan(&dbPercent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Segment{}, fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentNotFound)
		}

		return models.Segment{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return models.Segment{Slug: slug, Percent: dbPercent}, nil
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
