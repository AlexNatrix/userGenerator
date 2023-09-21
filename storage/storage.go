package storage

import (
	"database/sql"
	"fmt"
	"time"
)

const ErrConstraintsViolation = "Not enought fields provided"

type Storage struct {
	db          *sql.DB
	insertStats []stats
}

type stats struct {
	date      time.Time
	batchSize int
	failcount int
	inserted  int
	errors    []error
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
	sts := make([]stats, 0)
	return &Storage{db: db, insertStats: sts}, nil
}


