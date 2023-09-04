package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

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

	var resID int64

	err = stmt.QueryRow().Scan(&resID)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resID, nil
}

func (s *Storage) GetUser(id int64) (int64, error) {
	const op = "storage.postgres.GetUser"

	stmt, err := s.db.Prepare("SELECT id FROM users WHERE id = $1")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var resID int64

	err = stmt.QueryRow(id).Scan(&resID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: execute statement: %w", op, storage.ErrUserNotFound)
		}

		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resID, nil
}

func (s *Storage) GetUserSegments(id int64) ([]string, error) {
	const op = "storage.postgres.GetUserSegments"

	userID, err := s.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}

	stmt, err := s.db.Prepare(`
		SELECT slug
		FROM segments
		JOIN users_segments ON users_segments.segment_id = segments.id
		WHERE users_segments.user_id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
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
	segmentsToAdd []string,
	segmentsToRemove []string,
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

	historyStmt, err := tx.Prepare("INSERT INTO users_segments_history(user_id, segment_id, operation) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for history: %w", op, err)
	}
	defer historyStmt.Close()

	// Add the segments to the user
	addStmt, err := tx.Prepare("INSERT INTO users_segments(user_id, segment_id) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for add: %w", op, err)
	}
	defer addStmt.Close()

	for _, slug := range segmentsToAdd {
		segment, err := s.GetSegment(slug)
		if err != nil {
			return fmt.Errorf("%s: get segment: %w", op, err)
		}

		_, err = addStmt.Exec(userID, segment.ID)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("%s: %w", op, storage.ErrUserSegmentExists)
			}

			return fmt.Errorf("%s: insert user segment: %w", op, err)
		}

		_, err = historyStmt.Exec(userID, segment.ID, "add")
		if err != nil {
			return fmt.Errorf("%s: insert user segment history: %w", op, err)
		}
	}

	// Remove the segments from the user
	rmStmt, err := tx.Prepare("DELETE FROM users_segments WHERE user_id = $1 AND segment_id = $2")
	if err != nil {
		return fmt.Errorf("%s: create prepared statement for remove: %w", op, err)
	}
	defer rmStmt.Close()

	for _, slug := range segmentsToRemove {
		segment, err := s.GetSegment(slug)
		if err != nil {
			return fmt.Errorf("%s: get segment: %w", op, err)
		}

		res, err := rmStmt.Exec(userID, segment.ID)
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

		_, err = historyStmt.Exec(userID, segment.ID, "remove")
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

func (s *Storage) GetUserSegmentsHistory(userID int64, period time.Time) ([][]string, error) {
	const op = "storage.postgres.GetUserSegmentsHistory"

	if _, err := s.GetUser(userID); err != nil {
		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}

	rows, err := s.db.Query(`
		SELECT segment_id, operation, created_at
		FROM users_segments_history
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM created_at) = $2
		AND EXTRACT(MONTH FROM created_at) = $3
		ORDER BY created_at DESC
	`, userID, period.Year(), period.Month())
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	report := [][]string{}

	for rows.Next() {
		var segmentID string
		var operation string
		var datetime time.Time
		if err := rows.Scan(&segmentID, &operation, &datetime); err != nil {
			return nil, fmt.Errorf("%s: scanning rows: %w", op, err)
		}
		row := []string{
			strconv.FormatInt(userID, 10),
			segmentID,
			operation,
			datetime.Format(time.DateTime),
		}
		report = append(report, row)
	}

	return report, nil
}
