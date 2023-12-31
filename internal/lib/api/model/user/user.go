package internal

import "fmt"

type User struct {
	*BaseUser
	*Enrichment
}

func NewUser() User {
	return User{
		&BaseUser{
			Name:       "",
			Surname:    "",
			Patronymic: "",
		},
		&Enrichment{
			Age:         0,
			Sex:         "",
			Nationality: "",
		},
	}
}

func (u User) String() string {
	return fmt.Sprintf("{ name:%s, surname:%s, sex:%s, age:%d, nationality:%s }",
		u.Name, u.Surname, u.Sex, u.Age, u.Nationality)
}

type Enrichment struct {
	Age         int    `json:"age"`
	Sex         string `json:"gender"`
	Nationality string `json:"country"`
}

type BaseUser struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

func (bu *BaseUser) Validate() bool{
	if bu.Name=="" || bu.Surname == ""{
		return false
	}
	return true
}

func (u *BaseUser) String() string {
	return fmt.Sprintf("{ name:%s, surname:%s, patronymic:%s }",
		u.Name, u.Surname,u.Patronymic)
}

type CountryArray struct {
	Country []Country `json:"country"`
}
type Country struct {
	CountryID string `json:"country_id"`
}
