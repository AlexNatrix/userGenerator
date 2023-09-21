package storage

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	models "usergenerator/internal/lib/api/model/user"

	sq "github.com/Masterminds/squirrel"
)

//TODO:config
var allowedForSeach = make(map[string]bool)

func build(val map[string][]string) sq.SelectBuilder {
	allowedForSeach["name"] = true
	allowedForSeach["age"] = true
	allowedForSeach["surname"] = true
	allowedForSeach["patronymic"] = true
	allowedForSeach["nationality"] = true
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
	if _,ok:=val["page"];ok{
		page,err:=strconv.Atoi(val["page"][0])
			if err!=nil{
				page=1
			}
		if _,ok=val["per_page"];ok{
			perpage,err:=strconv.Atoi(val["per_page"][0])
			if err!=nil{
				perpage=100
			}
			sqb=sqb.Limit(uint64(page)).Offset(uint64(perpage)*uint64(page))
		}else{
			sqb=sqb.Limit(uint64(page))
		}
	}
	return sqb.PlaceholderFormat(sq.Dollar)
}

func (s *Storage) GetUsers(userQuery map[string][]string) ([]models.User, error) {
	const op = "storage.UpdateUser"
	sql, args, err := build(userQuery).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	data, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	result := make([]models.User, 0)
	for data.Next() {
		var id int64
		u := models.NewUser()
		err := data.Scan(&id, &u.Name, &u.Surname, &u.Patronymic, &u.Age, &u.Sex, &u.Nationality)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, u)
	}
	return result, nil
}
