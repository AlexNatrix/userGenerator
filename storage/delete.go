package storage

import "fmt"

func (s *Storage) DeleteUser(userID int64) (int64, error) {
	const op = "storage.DeleteUser"
	stmt, err := s.db.Prepare(`DELETE FROM users WHERE id = $1 RETURNING id;`)
	if err != nil {
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	var id int64
	err = stmt.QueryRow(userID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	stmt.Close()
	return id, nil
}
