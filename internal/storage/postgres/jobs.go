package postgres

import (
	"context"
	"fmt"
)

func (s *Storage) RemoveExpiredUsersSegments(ctx context.Context) (int64, error) {
	const op = "storage.postgres.RemoveExpiredUsersSegments"

	res, err := s.pool.Exec(ctx, "DELETE FROM users_segments WHERE expire_at < NOW()")
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return res.RowsAffected(), nil
}
