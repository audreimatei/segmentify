package postgres

import (
	"fmt"
)

func (s *Storage) RemoveExpiredUsersSegments() (int64, error) {
	const op = "storage.postgres.RemoveExpiredUsersSegments"

	res, err := s.db.Exec("DELETE FROM users_segments WHERE expire_at < NOW()")
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: get rows affected after delete: %w", op, err)
	}

	return rowsAffected, nil
}
