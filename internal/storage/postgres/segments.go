package postgres

import (
	"context"
	"errors"
	"fmt"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateSegment(ctx context.Context, segment models.Segment) (models.Segment, error) {
	const op = "storage.postgres.CreateSegment"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Segment{}, fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
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
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&usersCount); err != nil {
			return models.Segment{}, fmt.Errorf("%s: count users: %w", op, err)
		}

		usersToAddCount := usersCount * segment.Percent / 100
		usersToAdd, err := s.GetRandomUsers(ctx, usersToAddCount)
		if err != nil {
			return models.Segment{}, fmt.Errorf("%s: get random users: %w", op, err)
		}

		for _, userID := range usersToAdd {
			if _, err = tx.Exec(ctx, `
				INSERT INTO users_segments(user_id, segment_slug, expire_at)
				VALUES($1, $2, $3)
			`, userID, segment.Slug, nil); err != nil {
				return models.Segment{}, fmt.Errorf("%s: insert user segment: %w", op, err)
			}

			if _, err = tx.Exec(ctx, `
				INSERT INTO users_segments_history(user_id, segment_slug, operation)
				VALUES($1, $2, $3)
			`, userID, segment.Slug, "add"); err != nil {
				return models.Segment{}, fmt.Errorf("%s: insert user segment history: %w", op, err)
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.Segment{}, fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return segment, nil
}

func (s *Storage) GetSegment(ctx context.Context, slug string) (models.Segment, error) {
	const op = "storage.postgres.GetSegment"

	var dbPercent int64

	if err := s.pool.QueryRow(ctx, `
		SELECT percent
		FROM segments WHERE slug = $1
	`, slug).Scan(&dbPercent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Segment{}, fmt.Errorf("%s: execute statement: %w", op, storage.ErrSegmentNotFound)
		}

		return models.Segment{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return models.Segment{Slug: slug, Percent: dbPercent}, nil
}

func (s *Storage) DeleteSegment(ctx context.Context, slug string) error {
	const op = "storage.postgres.DeleteSegment"

	res, err := s.pool.Exec(ctx, `
		DELETE FROM segments
		WHERE slug = $1
	`, slug)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSegmentNotFound)
	}

	return nil
}
