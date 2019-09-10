package model

type User struct {
	Id     string `json:"id" gorm:"PRIMARY_KEY"`
	Grade  string `json:"grade"`
	Class  string `json:"class"`
	Number string `json:"number"`
	Name   string `json:"name"`
}
