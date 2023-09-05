package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateUser() (int64, error) {
	const op = "storage.postgres.CreateUser"

	stmt, err := s.db.Prepare("INSERT INTO users DEFAULT VALUES RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var dbID int64

	err = stmt.QueryRow().Scan(&dbID)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return dbID, nil
}

func (s *Storage) GetUser(id int64) (int64, error) {
	const op = "storage.postgres.GetUser"

	stmt, err := s.db.Prepare("SELECT id FROM users WHERE id = $1")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var dbID int64

	err = stmt.QueryRow(id).Scan(&dbID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: execute statement: %w", op, storage.ErrUserNotFound)
		}

		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return dbID, nil
}

func (s *Storage) GetUserSegments(id int64) ([]string, error) {
	const op = "storage.postgres.GetUserSegments"

	dbID, err := s.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}

	stmt, err := s.db.Prepare(`
		SELECT segment_slug
		FROM users_segments
		WHERE users_segments.user_id = $1
		AND (
			users_segments.expire_at IS NULL
			OR users_segments.expire_at > NOW()
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(dbID)
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	segments := []string{}

	for rows.Next() {
		var segment string
		if err := rows.Scan(&segment); err != nil {
			return nil, fmt.Errorf("%s: scanning rows: %w", op, err)
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

func (s *Storage) UpdateUserSegments(
	id int64,
	segmentsToAdd []models.SegmentToAdd,
	segmentsToRemove []models.SegmentToRemove,
) error {
	const op = "storage.postgres.UpdateUserSegments"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	userID, err := s.GetUser(id)
	if err != nil {
		return fmt.Errorf("%s: get user: %w", op, err)
	}

	historyStmt, err := tx.Prepare("INSERT INTO users_segments_history(user_id, segment_slug, operation) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for history: %w", op, err)
	}
	defer historyStmt.Close()

	// Add the segments to the user
	addStmt, err := tx.Prepare("INSERT INTO users_segments(user_id, segment_slug, expire_at) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for add: %w", op, err)
	}
	defer addStmt.Close()

	for _, segmentToAdd := range segmentsToAdd {
		segmentSlug, err := s.GetSegment(segmentToAdd.Slug)
		if err != nil {
			return fmt.Errorf("%s: get segment: %w", op, err)
		}

		if segmentToAdd.ExpireAt.IsZero() {
			_, err = addStmt.Exec(userID, segmentSlug, nil)
		} else {
			_, err = addStmt.Exec(userID, segmentSlug, segmentToAdd.ExpireAt)
		}
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("%s: %w", op, storage.ErrUserSegmentExists)
			}

			return fmt.Errorf("%s: insert user segment: %w", op, err)
		}

		_, err = historyStmt.Exec(userID, segmentSlug, "add")
		if err != nil {
			return fmt.Errorf("%s: insert user segment history: %w", op, err)
		}
	}

	// Remove the segments from the user
	rmStmt, err := tx.Prepare("DELETE FROM users_segments WHERE user_id = $1 AND segment_slug = $2")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for remove: %w", op, err)
	}
	defer rmStmt.Close()

	for _, segmentToRemove := range segmentsToRemove {
		segmentSlug, err := s.GetSegment(segmentToRemove.Slug)
		if err != nil {
			return fmt.Errorf("%s: get segment: %w", op, err)
		}

		res, err := rmStmt.Exec(userID, segmentSlug)
		if err != nil {
			return fmt.Errorf("%s: delete user segment: %w", op, err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: get rows affected after delete user segment: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, storage.ErrUserSegmentNotFound)
		}

		_, err = historyStmt.Exec(userID, segmentSlug, "remove")
		if err != nil {
			return fmt.Errorf("%s: insert user segment history: %w", op, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserSegmentsHistory(id int64, period time.Time) ([][]string, error) {
	const op = "storage.postgres.GetUserSegmentsHistory"

	if _, err := s.GetUser(id); err != nil {
		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}

	rows, err := s.db.Query(`
		SELECT user_id, segment_slug, operation, created_at
		FROM users_segments_history
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM created_at) = $2
		AND EXTRACT(MONTH FROM created_at) = $3
	`, id, period.Year(), period.Month())
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	report := [][]string{}

	for rows.Next() {
		var userID int64
		var segmentSlug string
		var operation string
		var created_at time.Time
		if err := rows.Scan(&userID, &segmentSlug, &operation, &created_at); err != nil {
			return nil, fmt.Errorf("%s: scanning rows: %w", op, err)
		}
		row := []string{
			strconv.FormatInt(userID, 10),
			segmentSlug,
			operation,
			created_at.Format(time.RFC3339),
		}
		report = append(report, row)
	}

	return report, nil
}
