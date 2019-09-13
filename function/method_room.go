package function

import (
	"Documents/attendance_book/server/src/model"

	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//Roomの新規作成
func InsertRoomData(w http.ResponseWriter, r *http.Request) {
	var room model.Room
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (dbGetOneLecture)")
	}
	defer db.Close()
	log.Printf("POST: InsertRoomData")
	decoder := json.NewDecoder(r.Body)
	error := decoder.Decode(&room)
	if error != nil {
		w.Write([]byte("json decode error" + error.Error() + "\n"))
	}
	//uuid生成
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return
	}
	uu := u.String()

	vars := mux.Vars(r)
	room.OwnerId = vars["id"]
	room.Id = uu

	//同一パスワードを弾く
	var check model.Check
	var checkRooms []model.Room
	db.Find(&checkRooms)
	for i := 0; i < len(checkRooms); i++ {
		if checkRooms[i].Password == room.Password {
			check.Value = false
			json.NewEncoder(w).Encode(check.Value)
			log.Printf("11!!!!!111\n")
			return
		}
	}

	db.Create(&model.Room{
		Id:          uu,
		OwnerId:     room.OwnerId,
		Title:       room.Title,
		Password:    room.Password,
		Description: room.Description,
	})

	check.Value = true
	json.NewEncoder(w).Encode(check.Value)
	/*
		debug
		fmt.Printf("room:%+v\n", room)
	*/
}

//所有しているroomの情報を得る
func GetOwnerRoomData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (GetRoomData)")
	}
	defer db.Close()
	log.Printf("GET: GetRoomData")
	vars := mux.Vars(r)
	owner_id := vars["id"]
	var rooms []model.Room
	db.Where(&model.Room{OwnerId: owner_id}).Find(&rooms)
	json.NewEncoder(w).Encode(rooms)
}

//ルームパスコードチェック
func CheckRoomPassword(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (CheckRoomPassword)")
	}
	defer db.Close()
	log.Printf("GET: CheckRoomPassword")
	vars := mux.Vars(r)
	user_id := vars["userid"]
	password := vars["password"]
	log.Printf("userid:" + user_id + "\npassword:" + password)
	var rooms []model.Room
	db.Find(&rooms)

	var check model.Check

	for i := 0; i < len(rooms); i++ {
		if rooms[i].Password == password {
			newRecordbool := InsertRelationUserRoom(user_id, rooms[i].Id)

			if newRecordbool == true {
				/*
					debug
					log.Printf(rooms[i].Id)
				*/
				check.Value = true
				json.NewEncoder(w).Encode(check)
				return
			} else {
				check.Value = false
				json.NewEncoder(w).Encode(check)
				return
			}
		}
	}
	log.Printf("not much room pass")
	check.Value = false
	json.NewEncoder(w).Encode(check)
}

func InsertRelationUserRoom(user_id string, room_id string) bool {
	//データベースを開く
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (InsertRelationUserRoom)")
	}
	defer db.Close()

	//uuid生成
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return false
	}
	uu := u.String()

	//roomのtitleを得るためデータベースに接続
	var room model.Room
	db.Where(&model.Room{Id: room_id}).First(&room)

	/*
		debug
		fmt.Printf("room:%+v\n", room)
	*/

	//同一のuser_idとroom_idが存在していないかチェック
	var checks []model.RelationUserRoom
	db.Where(&model.RelationUserRoom{UserId: user_id, RoomId: room_id}).Find(&checks)

	/*
		debug
		for i := 0; i < len(checks); i++ {
			fmt.Printf("checks:%+v\n", checks[i])
		}
	*/

	//nilのときはokそれ以外はuser_idが等しいことを利用して弾く
	for i := 0; i < len(checks); i++ {
		if checks[i].UserId == user_id {
			return false
		}
	}

	//dataの作成
	data := model.RelationUserRoom{
		Id:        uu,
		RoomTitle: room.Title,
		RoomId:    room_id,
		UserId:    user_id,
	}
	/*
		debug
		fmt.Printf("data:%+v\n", data)
	*/

	db.Create(&data)
	return true
}

func GetMemberRoomData(w http.ResponseWriter, r *http.Request) {
	//データベースを開く
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (InsertRelationUserRoom)")
	}
	defer db.Close()
	//urlからuser_idを取得する
	vars := mux.Vars(r)
	user_id := vars["id"]

	//user_idに関連付けられたrooms_idをデータベースから取得する
	var rooms_id []model.RelationUserRoom
	db.Where(&model.RelationUserRoom{UserId: user_id}).Find(&rooms_id)

	//rooms_idをもとにroomのデータを構造体配列roomsに格納していく
	var rooms []model.Room
	for i := 0; i < len(rooms_id); i++ {
		var room model.Room
		db.Where(&model.Room{Id: rooms_id[i].RoomId}).First(&room)
		rooms = append(rooms, room)
	}

	/*
		debug

		for i := 0; i < len(rooms); i++ {
			fmt.Printf("user admin room:%+v\n", rooms[i])
		}
	*/

	json.NewEncoder(w).Encode(rooms)
}
