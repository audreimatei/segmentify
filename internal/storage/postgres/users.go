package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateUser(ctx context.Context) (int64, error) {
	fail := func(msg string, err error) (int64, error) {
		return 0, fmt.Errorf("storage.postgres.CreateUser: %s: %w", msg, err)
	}

	var dbID int64

	if err := s.pool.QueryRow(ctx, `
		INSERT INTO users
		DEFAULT VALUES
		RETURNING id
	`).Scan(&dbID); err != nil {
		return fail("insert user with returning", err)
	}

	return dbID, nil
}

func (s *Storage) GetUser(ctx context.Context, id int64) (int64, error) {
	fail := func(msg string, err error) (int64, error) {
		return 0, fmt.Errorf("storage.postgres.GetUser: %s: %w", msg, err)
	}

	var dbID int64

	if err := s.pool.QueryRow(ctx, `
		SELECT id
		FROM users
		WHERE id = $1
	`, id).Scan(&dbID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fail("query user", storage.ErrUserNotFound)
		}
		return fail("query user", err)
	}

	return dbID, nil
}

func (s *Storage) GetRandomUsers(ctx context.Context, usersCount int64) ([]int64, error) {
	fail := func(msg string, err error) ([]int64, error) {
		return []int64{}, fmt.Errorf("storage.postgres.GetRandomUsers: %s: %w", msg, err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM users
		ORDER BY RANDOM()
		LIMIT $1
	`, usersCount)
	if err != nil {
		return fail("query users", err)
	}
	defer rows.Close()

	users := make([]int64, 0, usersCount)

	for rows.Next() {
		var user int64
		if err = rows.Scan(&user); err != nil {
			return fail("scan users", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *Storage) GetUserSegments(ctx context.Context, id int64) ([]string, error) {
	fail := func(msg string, err error) ([]string, error) {
		return []string{}, fmt.Errorf("storage.postgres.GetUserSegments: %s: %w", msg, err)
	}

	dbID, err := s.GetUser(ctx, id)
	if err != nil {
		return fail("get user", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT segment_slug
		FROM users_segments
		WHERE users_segments.user_id = $1
		AND (
			users_segments.expire_at IS NULL
			OR users_segments.expire_at > NOW()
		)
	`, dbID)
	if err != nil {
		return fail("query user segments", err)
	}
	defer rows.Close()

	segments := []string{}

	for rows.Next() {
		var segment string
		if err = rows.Scan(&segment); err != nil {
			return fail("scan user segments", err)
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

func (s *Storage) UpdateUserSegments(
	ctx context.Context,
	id int64,
	segmentsToAdd []models.SegmentToAdd,
	segmentsToRemove []models.SegmentToRemove,
) error {
	fail := func(msg string, err error) error {
		return fmt.Errorf("storage.postgres.UpdateUserSegments: %s: %w", msg, err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fail("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	userID, err := s.GetUser(ctx, id)
	if err != nil {
		return fail("get user", err)
	}

	// Add the segments to the user
	for _, segmentToAdd := range segmentsToAdd {
		segment, err := s.GetSegment(ctx, segmentToAdd.Slug)
		if err != nil {
			return fail("get segment to add", err)
		}

		expireAt := &segmentToAdd.ExpireAt
		if segmentToAdd.ExpireAt.IsZero() {
			expireAt = nil
		}

		if _, err = s.pool.Exec(ctx, `
				INSERT INTO users_segments(user_id, segment_slug, expire_at)
				VALUES($1, $2, $3)
			`, userID, segment.Slug, expireAt); err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				return fail("insert user segment", storage.ErrUserSegmentExists)
			}
			return fail("insert user segment", err)
		}

		if _, err = s.pool.Exec(ctx, `
			INSERT INTO users_segments_history(user_id, segment_slug, operation)
			VALUES($1, $2, $3)
		`, userID, segment.Slug, "add"); err != nil {
			return fail("insert user segment history, add", err)
		}
	}

	// Remove the segments from the user
	for _, segmentToRemove := range segmentsToRemove {
		segment, err := s.GetSegment(ctx, segmentToRemove.Slug)
		if err != nil {
			return fail("get segment to remove", err)
		}

		res, err := s.pool.Exec(ctx, `
			DELETE FROM users_segments
			WHERE user_id = $1
			AND segment_slug = $2
		`, userID, segment.Slug)
		if err != nil {
			return fail("delete user segment", err)
		}

		if res.RowsAffected() == 0 {
			return fail("rows affected", storage.ErrUserSegmentNotFound)
		}

		_, err = s.pool.Exec(ctx, `
			INSERT INTO users_segments_history(user_id, segment_slug, operation)
			VALUES($1, $2, $3)
		`, userID, segment.Slug, "remove")
		if err != nil {
			return fail("insert user segment history, remove", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fail("commit transaction", err)
	}

	return nil
}

func (s *Storage) GetUserSegmentsHistory(
	ctx context.Context,
	id int64,
	period time.Time,
) ([][]string, error) {
	fail := func(msg string, err error) ([][]string, error) {
		return [][]string{}, fmt.Errorf("storage.postgres.GetUserSegmentsHistory: %s: %w", msg, err)
	}

	if _, err := s.GetUser(ctx, id); err != nil {
		return fail("get user", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT user_id, segment_slug, operation, created_at
		FROM users_segments_history
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM created_at) = $2
		AND EXTRACT(MONTH FROM created_at) = $3
	`, id, period.Year(), period.Month())
	if err != nil {
		return fail("query history", err)
	}
	defer rows.Close()

	report := [][]string{}

	for rows.Next() {
		var userID int64
		var segmentSlug string
		var operation string
		var created_at time.Time
		if err := rows.Scan(&userID, &segmentSlug, &operation, &created_at); err != nil {
			return fail("scan history", err)
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
