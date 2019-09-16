package function

import (
	"Documents/attendance_book/server/src/model"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//_ "github.com/mattn/go-sqlite3"

var DatabaseName string
var DatabaseUrl string

//データベース初期化
func DbInit() {
	//データベース関連
	//DatabaseUrl = "test.sqlite3"
	//DatabaseName = "sqlite3"
	DatabaseUrl = os.Getenv("DATABASE_URL")
	DatabaseName = "postgres"

	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database!（dbInit）")
	}
	//残りのモデルはまだ入れてない。
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Room{})
	db.AutoMigrate(&model.List{})
	db.AutoMigrate(&model.RelationUserRoom{})
	db.AutoMigrate(&model.RelationUserList{})
	db.AutoMigrate(&model.PasswordList{})
	defer db.Close()
}

//ユーザーデータの新規登録
func InsertUserData(w http.ResponseWriter, r *http.Request) {
	var User model.User

	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database!（dbInsert）")
	}
	defer db.Close()

	log.Printf("POST: InsertUserData")
	vars := mux.Vars(r)
	decoder := json.NewDecoder(r.Body)
	error := decoder.Decode(&User)
	if error != nil {
		w.Write([]byte("json decode error" + error.Error() + "\n"))
	}
	User.Id = vars["id"]
	db.Create(&model.User{
		Id:     User.Id,
		Grade:  User.Grade,
		Class:  User.Class,
		Number: User.Number,
		Name:   User.Name,
	})
	fmt.Printf("user:%+v\n", User)
}

func EditUserData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database!（dbInsert）")
	}
	defer db.Close()

	decoder := json.NewDecoder(r.Body)

	var newuser model.User
	error := decoder.Decode(&newuser)
	if error != nil {
		w.Write([]byte("json decode error" + error.Error() + "\n"))
	}
	vars := mux.Vars(r)
	newuser.Id = vars["id"]

	var user model.User
	db.Where("Id = ?", newuser.Id).First(&user)
	user.Id = newuser.Id
	user.Class = newuser.Class
	user.Grade = newuser.Grade
	user.Number = newuser.Number
	user.Name = newuser.Name
	db.Save(&user)
}

//あるidのユーザーデータの取得
func GetUserData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (GetUserData)")
	}
	defer db.Close()
	log.Printf("GET: GetUserData")
	vars := mux.Vars(r)
	id := vars["id"]
	var user model.User
	db.Where("Id = ?", id).First(&user)
	json.NewEncoder(w).Encode(user)
}

func Private(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET: Private")
	w.Write([]byte("hello private!\n"))
}

func GetAllRoomMember(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (GetAllRoomMember)")
	}
	defer db.Close()
	log.Printf("GET: GetAllRoomMember")
	vars := mux.Vars(r)
	roomId := vars["id"]
	var usersId []model.RelationUserRoom
	db.Where(&model.RelationUserRoom{RoomId: roomId}).Find(&usersId)

	var users []model.User
	for i := 0; i < len(usersId); i++ {
		var user model.User
		db.Where(&model.User{Id: usersId[i].UserId}).First(&user)
		users = append(users, user)
	}
	log.Printf("allroomuser: %v", users)
	json.NewEncoder(w).Encode(users)
}
