package internal

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	models "main/internal/lib/api/model/user"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const ErrConstraintsViolation = "Not enought fields provided"

type Storage struct {
	db *sql.DB
}

func New(StoragePath string) (*Storage, error) {
	const op = "storage.pg.New"
	db, err := sql.Open("postgres", StoragePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			name VARCHAR (40)  NOT NULL CONSTRAINT name CHECK(length(name)>0),
			surname VARCHAR (40) NOT NULL CONSTRAINT surname CHECK(length(surname)>0),
			patronymic TEXT,
			age numeric NOT NULL CHECK(age > 0),
			sex VARCHAR (6) NOT NULL CONSTRAINT sex CHECK(length(sex )>0),
			nationality TEXT NOT NULL CONSTRAINT nationality CHECK(length(nationality)>0)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(log *slog.Logger, users ...models.User) ([]int64, error) {
	const op = "storage.SaveUser"
	stmt, err := s.db.Prepare(`INSERT INTO users(name, surname, patronymic, age, sex, nationality)
	VALUES($1, $2, $3, $4, $5, $6) RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ids := make([]int64, len(users))
	counter := 0
	for i, user := range users {
		err = stmt.QueryRow(
			user.Name, user.Surname,
			user.Patronymic, user.Age,
			user.Sex, user.Nationality).Scan(&ids[i])
		if err != nil {
			counter++
			if pgErr, ok := err.(*pq.Error); ok && (pgErr.Code == "23514" || pgErr.Code == "23505") {
				log.Error("Failed to proceed in DB,failed constraits", fmt.Errorf("%s : %s", op, ErrConstraintsViolation))
			}

			log.Error("Failed to proceed in DB,failed constraits", fmt.Errorf("%s : %w", op, err))
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && (pgErr.Code == "23514" || pgErr.Code == "23505") {
			return ids, fmt.Errorf("%s : %s", op, ErrConstraintsViolation)
		}

		return ids, fmt.Errorf("%s : %w", op, err)
	}
	stmt.Close()
	if counter >= len(ids) {
		return nil, fmt.Errorf("%s : %w", op, fmt.Errorf("Whole batch failed"))
	}
	return ids, nil
}

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

var allowedForSeach = make(map[string]bool)

func Build(val map[string][]string) sq.SelectBuilder {
	allowedForSeach["name"] = true
	allowedForSeach["age"] = true
	allowedForSeach["sex"] = true
	sqb := sq.Select("*").From("users")
	for k := range val {
		if allowedForSeach[k] {
			q := strings.Split(val[k][0], "~")
			var op string
			var v string
			if len(q) > 1 {
				op = q[0]
				v = q[1]
			} else {
				v = q[0]
				op = ""
			}
			fmt.Println(op, "op", q, len(q), "q")
			switch op {
			case "gt":
				sqb = sqb.Where(sq.Gt{k: v})
			case "lt":
				sqb = sqb.Where(sq.Lt{k: v})
			case "neq":
				sqb = sqb.Where(sq.NotEq{k: v})
			default:
				sqb = sqb.Where(sq.Eq{k: v})
			}
		}
	}
	fmt.Println(sqb.ToSql())
	return sqb.PlaceholderFormat(sq.Dollar)
}

func (s *Storage) GetUsers(userQuery map[string][]string) ([]models.User, error) {
	const op = "storage.UpdateUser"
	sql, args, err := Build(userQuery).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	fmt.Println(args...)
	data, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	result := make([]models.User, 0)
	for data.Next() {
		var id int64
		u := models.New()
		err := data.Scan(&id, &u.Name, &u.Surname, &u.Patronymic, &u.Age, &u.Sex, &u.Nationality)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, u)
	}
	fmt.Println(result, "lol")
	return result, nil
}

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
	fmt.Println(sql)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	fmt.Println(args...)
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
