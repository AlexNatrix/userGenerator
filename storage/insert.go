package storage

import (
	"fmt"
	"time"
	models "usergenerator/internal/lib/api/model/user"

	"github.com/lib/pq"
)

func (s *Storage) InsertUsers(users ...models.User) ([]int64, error) {
	const op = "storage.SaveUser"
	if len(users) == 0 {
		return nil, nil
	}
	stmt, err := s.db.Prepare(`INSERT INTO users(name, surname, patronymic, age, sex, nationality)
	VALUES($1, $2, $3, $4, $5, $6) RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	ids := make([]int64, len(users))
	batchSize := len(users)
	failcount := 0
	errors := make([]error, 0)
	for i, user := range users {
		err = stmt.QueryRow(
			user.Name, user.Surname,
			user.Patronymic, user.Age,
			user.Sex, user.Nationality).Scan(&ids[i])
		if err != nil {
			failcount++
			if pgErr, ok := err.(*pq.Error); ok && (pgErr.Code == "23514" || pgErr.Code == "23505") {
				errors = append(errors, fmt.Errorf("%s : %s", op, ErrConstraintsViolation))
			}
			errors = append(errors, fmt.Errorf("%s : %s", op, ErrConstraintsViolation))
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		errors = append(errors, fmt.Errorf("%s LAST EXEC (NOT AN ERROR) for stats: %w", op, err))
	}
	stmt.Close()
	if failcount >= len(ids) {
		return nil, fmt.Errorf("%s : %w", op, fmt.Errorf("whole batch failed or empty"))
	}
	if len(s.insertStats) > 20 {
		s.insertStats = s.insertStats[1:]
		s.insertStats = append(s.insertStats, stats{
			date:      time.Now().UTC().In(time.Local),
			batchSize: batchSize,
			failcount: failcount,
			inserted:  batchSize - failcount,
			errors:    errors,
		})
	}
	return ids, nil
}
