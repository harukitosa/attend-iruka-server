package function

import (
	"Documents/attendance_book/server/src/model"

	"encoding/json"
	"fmt"
	"log"
    "os"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DatabaseName string
var DatabaseUrl string

//データベース初期化
func DbInit() {
	//データベース関連
    // DatabaseUrl = "test.sqlite3"
	//DatabaseName = "sqlite3"
    DatabaseUrl := os.Getenv("DATABASE_URL")
    DatabaseName := "postgres"

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
