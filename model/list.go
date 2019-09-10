package model

import "time"

type List struct {
	Id     string `gorm:"PRIMARY_KEY" json:"id"`
	RoomId string `json:"room_id"`
	Title  string `json:"title"`
}

type PasswordList struct {
	ListId   string    `gorm:"PRIMARY_KEY" json:"id"`
    //もしかしたらuserIdは不要
	UserId   string    `json:"user_id"`
	Password string    `json:"password"`
	Time     time.Time `json:"time"`
}

type RelationUserRoom struct {
    RoomTitle string `json:"room_title"`
 	RoomId string `json:"room_id"`
   	UserId string `json:"user_id"`
    Id string `gorm:"PRIMARY_KEY" json:"id"`
}

type RelationUserList struct {
    Id string `gorm:"PRIMARY_KEY" json:"id"`    
    UserId string `json:"user_id"`
	ListId   string    `json:"list_id"`
}
