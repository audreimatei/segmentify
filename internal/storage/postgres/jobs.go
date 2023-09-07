package postgres

import (
	"context"
	"fmt"
)

func (s *Storage) DeleteExpiredUsersSegments(ctx context.Context) (int64, error) {
	fail := func(msg string, err error) (int64, error) {
		return 0, fmt.Errorf("storage.postgres.DeleteExpiredUsersSegments: %s: %w", msg, err)
	}

	res, err := s.pool.Exec(ctx, `
		DELETE FROM users_segments
		WHERE expire_at < NOW()
	`)
	if err != nil {
		return fail("delete users segments", err)
	}

	return res.RowsAffected(), nil
}
