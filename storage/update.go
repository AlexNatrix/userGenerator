package storage

import (
	"fmt"
	models "usergenerator/internal/lib/api/model/user"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

func (s *Storage) UpdateUser(userID int64, user models.User) error {
	const op = "storage.UpdateUser"
	sql, args, err := sq.Update("users").
		Set("name", user.Name).
		Set("surname", user.Surname).
		Set("patronymic ", user.Patronymic).
		Set("age ", user.Age).
		Set("sex", user.Sex).
		Set("nationality ", user.Nationality).
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	_, err = stmt.Exec(args...)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && (pgErr.Code == "23514" || pgErr.Code == "23505") {
			return fmt.Errorf("%s : %s", op, ErrConstraintsViolation)
		}

		return fmt.Errorf("%s : %w", op, err)
	}
	stmt.Close()
	return nil
}
