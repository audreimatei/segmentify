package postgres

import (
	"context"
	"errors"
	"fmt"
	"segmentify/internal/models"
	"segmentify/internal/storage"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateSegment(ctx context.Context, segment models.Segment) (models.Segment, error) {
	fail := func(msg string, err error) (models.Segment, error) {
		return models.Segment{}, fmt.Errorf("storage.postgres.CreateSegment: %s: %w", msg, err)
	}
	failRowsAffected := func(msg, expected, got string) (models.Segment, error) {
		return fail(msg, fmt.Errorf("not enough rows affected; expected: %s, got: %s", expected, got))
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fail("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `
		INSERT INTO segments(slug, percent)
		VALUES($1, $2)
	`, segment.Slug, segment.Percent); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return fail("insert segment", &storage.ErrSegmentExists{Slug: segment.Slug})
		}
		return fail("insert segment", err)
	}

	if segment.Percent > 0 {
		var usersCount int64
		if err = tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
		`).Scan(&usersCount); err != nil {
			return fail("count users", err)
		}

		usersToAddCount := usersCount * segment.Percent / 100
		usersToAdd, err := s.GetRandomUsers(ctx, usersToAddCount)
		if err != nil {
			return fail("get random users", err)
		}

		rowsAffected, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"users_segments"},
			[]string{"user_id", "segment_slug", "expire_at"},
			pgx.CopyFromSlice(len(usersToAdd), func(i int) ([]any, error) {
				return []any{usersToAdd[i], segment.Slug, nil}, nil
			}),
		)
		if err != nil {
			return fail("insert users segments", err)
		}
		if rowsAffected != usersToAddCount {
			return failRowsAffected(
				"insert users segments",
				strconv.FormatInt(usersToAddCount, 10),
				strconv.FormatInt(rowsAffected, 10),
			)
		}

		rowsAffected, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"users_segments_history"},
			[]string{"user_id", "segment_slug", "operation"},
			pgx.CopyFromSlice(len(usersToAdd), func(i int) ([]any, error) {
				return []any{usersToAdd[i], segment.Slug, "add"}, nil
			}),
		)
		if err != nil {
			return fail("insert users segments history", err)
		}
		if rowsAffected != usersToAddCount {
			return failRowsAffected(
				"insert users segments history",
				strconv.FormatInt(usersToAddCount, 10),
				strconv.FormatInt(rowsAffected, 10),
			)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fail("commit transaction", err)
	}

	return segment, nil
}

func (s *Storage) GetSegment(ctx context.Context, slug string) (models.Segment, error) {
	fail := func(msg string, err error) (models.Segment, error) {
		return models.Segment{}, fmt.Errorf("storage.postgres.GetSegment: %s: %w", msg, err)
	}

	var dbPercent int64

	if err := s.pool.QueryRow(ctx, `
		SELECT percent
		FROM segments
		WHERE slug = $1
	`, slug).Scan(&dbPercent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fail("query segment", &storage.ErrSegmentNotFound{Slug: slug})
		}
		return fail("query segment", err)
	}

	return models.Segment{Slug: slug, Percent: dbPercent}, nil
}

func (s *Storage) DeleteSegment(ctx context.Context, slug string) error {
	fail := func(msg string, err error) error {
		return fmt.Errorf("storage.postgres.DeleteSegment: %s: %w", msg, err)
	}

	res, err := s.pool.Exec(ctx, `
		DELETE FROM segments
		WHERE slug = $1
	`, slug)
	if err != nil {
		return fail("delete segment", err)
	}

	if res.RowsAffected() == 0 {
		return fail("rows affected", &storage.ErrSegmentNotFound{Slug: slug})
	}

	return nil
}
