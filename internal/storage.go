package internal

import (
	"database/sql"
	"fmt"
	models "main/internal/lib/api/model/user"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const ErrConstraintsViolation = "Not enought fields provided"



type Storage struct {
	db *sql.DB
}



func New(StoragePath string) (*Storage,error){
	const op = "storage.pg.New"
	db,err := sql.Open("postgres",StoragePath)
	if err!= nil{
		return nil, fmt.Errorf("%s: %w", op,err)
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
	if err != nil{
		return nil,fmt.Errorf("%s: %w", op,err)
	}
	_,err = stmt.Exec()
	if err!=nil{
		return nil,fmt.Errorf("%s: %w", op,err)
	}
	return &Storage{db:db},nil
}


func (s *Storage) SaveUser(user models.User) (int64, error){
	const op = "storage.SaveUser"
	stmt, err := s.db.Prepare(`INSERT INTO users(name, surname, patronymic, age, sex, nationality)
	VALUES($1, $2, $3, $4, $5, $6) RETURNING id`)
	if err!=nil{
		return 0, fmt.Errorf("%s : %w",op,err)
	}
	var id int64
	err = stmt.QueryRow(
		user.Name,user.Surname,
		user.Patronymic,user.Age,
		user.Sex,user.Nationality).Scan(&id)
	if err!=nil{
		if pgErr,ok:= err.(*pq.Error);ok && (pgErr.Code=="23514" || pgErr.Code=="23505"){
			return 0, fmt.Errorf("%s : %s",op,ErrConstraintsViolation)
		}

		return 0, fmt.Errorf("%s : %w",op,err)
	}
	stmt.Close()
	return id, nil
}


func (s *Storage) DeleteUser(userID int64) (int64, error){
	const op = "storage.DeleteUser"
	stmt, err := s.db.Prepare(`DELETE FROM users WHERE id = $1 RETURNING id;`)
	if err!=nil{
		return 0, fmt.Errorf("%s : %w",op,err)
	}
	var id int64
	err = stmt.QueryRow(userID).Scan(&id)
	if err!=nil{
		return 0, fmt.Errorf("%s : %w",op,err)
	}
	stmt.Close()
	return id, nil
}





func (s *Storage) UpdateUser(userID int64, user models.User) (error){
	const op = "storage.UpdateUser"
	sql , args, err :=  sq.Update("users").
                        Set("name", user.Name).
                        Set("surname",  user.Surname).
						Set("patronymic ",  user.Patronymic).
						Set("age ",  user.Age).
						Set("sex",  user.Sex).
						Set("nationality ",  user.Nationality).
                        Where(sq.Eq{"id": userID}).
                        PlaceholderFormat(sq.Dollar).ToSql()
	fmt.Println(sql)
	if err!=nil{
		return fmt.Errorf("%s : %w",op,err)
	}					
	stmt,err := s.db.Prepare(sql)
	if err!=nil{
		return  fmt.Errorf("%s : %w",op,err)
	}
	fmt.Println(args...)
	_,err = stmt.Exec(args...)
	if err!=nil{
		if pgErr,ok:= err.(*pq.Error);ok && (pgErr.Code=="23514" || pgErr.Code=="23505"){
			return fmt.Errorf("%s : %s",op,ErrConstraintsViolation)
		}

		return fmt.Errorf("%s : %w",op,err)
	}
	stmt.Close()
	return nil
}

