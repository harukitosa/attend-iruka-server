package function

/*
主にlistのデータをあつかう関数を格納している。
InsertListData: listの新規作成
GetOwnerListData: RoomidからそのRoom内のlistをすべて取得している。
CheckListPassword:
*/

import (
	"Documents/attendance_book/server/src/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//listの新規作成
func InsertListData(w http.ResponseWriter, r *http.Request) {
	var list model.List
	var password model.PasswordList

	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (dbGetOneLecture)")
	}
	defer db.Close()
	log.Printf("POST: InsertListData")
	decoder := json.NewDecoder(r.Body)
	error := decoder.Decode(&password)
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
	list.Id = uu

	//時刻
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	s := t.Format(layout)

	vars := mux.Vars(r)
	list.RoomId = vars["id"]
	list.Title = s
	password.Time = t
	password.ListId = list.Id

	db.Create(&model.List{
		Id:     list.Id,
		RoomId: list.RoomId,
		Title:  list.Title,
	})

	db.Create(&model.PasswordList{
		ListId:   list.Id,
		UserId:   password.UserId,
		Password: password.Password,
		Time:     t,
	})
	var all []model.PasswordList
	db.Find(&all)
	fmt.Printf("all pass list:%+v\n", all)
	fmt.Printf("list:%+v\n", list)
	fmt.Printf("passlist:%+v\n", password)
}

//所有しているルームのリストデータ取り出し
func GetOwnerListData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (GetOwnerListData)")
	}
	defer db.Close()
	log.Printf("GET: GetOwnerListData")
	vars := mux.Vars(r)
	room_id := vars["id"]
	var lists []model.List
	db.Where(&model.List{RoomId: room_id}).Find(&lists)
	json.NewEncoder(w).Encode(lists)
}

/*
listのパスワードを取り出す
*/
func CheckListPassword(w http.ResponseWriter, r *http.Request) {

	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (CheckListPassword)")
	}
	defer db.Close()

	log.Printf("GET: CheckListPassword")
	vars := mux.Vars(r)
	password := vars["password"]
	room_id := vars["roomid"]
	user_id := vars["userid"]

	/*
		    debug

			log.Println("password:" + password)
			log.Println("room_id:" + room_id)
		    log.Println("user_id:" + user_id)
	*/
	var check model.Check

	var lists []model.List
	db.Where(&model.List{RoomId: room_id}).Find(&lists)

	var passwordLists []model.PasswordList
	for i := 0; i < len(lists); i++ {
		var passlist model.PasswordList
		db.Where(&model.PasswordList{ListId: lists[i].Id}).First(&passlist)
		/*
		   debug
		   fmt.Printf("pass:%+v\n", passlist)
		*/

		/*
		   取り出したroom内のlist_passwordをすべて当たってpassListに格納していく。
		   その後、passwordが一致したらpasswordListsに格納する。
		   !!!!!warning!!!!!
		   同一Room内に同一のパスワードを持つlistが存在した場合、うまく作動しない。
		*/
		if passlist.Password == password {
			passwordLists = append(passwordLists, passlist)
		}
	}

	/*
		    debug
			for i := 0; i < len(passwordLists); i++ {
				fmt.Printf("passwordList:%+v\n", passwordLists[i])
		    }
	*/

	/*
	   passwordListsの長さがゼロ、つまりlistが存在しないかpasswordの一致するリストがなかった場合falseを返す
	*/
	if len(passwordLists) == 0 {
		check.Value = false
		json.NewEncoder(w).Encode(check)
		return
	}

	/*
	   以下、タイムアウトの実装
	   herokuの時刻をJapanにしておく
	*/
	now := time.Now()
	back := passwordLists[0].Time
	timepass := back.Add(15 * time.Minute)

	/*
	   時間が十五分以内のときtrueを返す、それ以外はfalseを返す。
	*/
	if now.Before(timepass) {
		check.Value = InsertRelationUserList(user_id, passwordLists[0].ListId)
		log.Println("time is ok")
		json.NewEncoder(w).Encode(check)
		return

	} else {
		check.Value = false
		log.Println("time is bas")
		json.NewEncoder(w).Encode(check)
		return
	}
}

func InsertRelationUserList(user_id string, list_id string) bool {
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

	//同一のuser_idとlist_idが存在していないかチェック
	var checks []model.RelationUserList
	db.Where(&model.RelationUserList{UserId: user_id, ListId: list_id}).Find(&checks)
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

	data := model.RelationUserList{
		Id:     uu,
		UserId: user_id,
		ListId: list_id,
	}
	/*
	   debug
	   fmt.Printf("data:%+v\n", data)
	*/
	db.Create(&data)
	return true
}

/*
出席している生徒全員の情報を返す関数
*/
func GetListMember(w http.ResponseWriter, r *http.Request) {
	//データベースを開く
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database! (InsertRelationUserRoom)")
	}
	defer db.Close()

	log.Printf("GET: GetListMember")
	vars := mux.Vars(r)
	list_id := vars["listid"]

	var lists []model.RelationUserList
	db.Where(&model.RelationUserList{ListId: list_id}).Find(&lists)

	/*
		    debug
			for i := 0; i < len(lists); i++ {
				fmt.Printf("checks:%+v\n", lists[i])
			}
	*/

	var users []model.User
	for i := 0; i < len(lists); i++ {
		var user model.User
		db.Where(&model.User{Id: lists[i].UserId}).First(&user)
		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}

/*
Roomに登録しているuser全員の情報を得ている
*/
func GetListAllMember(w http.ResponseWriter, r *http.Request) {
	//データベースを開く
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database GetListAllMember")
	}
	defer db.Close()
	log.Printf("GET: GetListAllMember")

	vars := mux.Vars(r)
	listID := vars["listid"]

	//以下ユーザー情報の全取得
	var num model.List
	db.Where(&model.List{Id: listID}).First(&num)
	var listRoom model.Room
	db.Where(&model.Room{Id: num.RoomId}).First(&listRoom)
	var allUsersId []model.RelationUserRoom
	db.Where(&model.RelationUserRoom{RoomId: listRoom.Id}).Find(&allUsersId)
	log.Printf("#####################\n%+v", allUsersId)

	var all_users []model.User
	for i := 0; i < len(allUsersId); i++ {
		var all_user model.User
		db.Where(&model.User{Id: allUsersId[i].UserId}).First(&all_user)
		all_users = append(all_users, all_user)
	}

	var lists []model.RelationUserList
	db.Where(&model.RelationUserList{ListId: listID}).Find(&lists)
	for i := 0; i < len(lists); i++ {
		fmt.Printf("checks:%+v\n", lists[i])
	}

	var users []model.User
	for i := 0; i < len(lists); i++ {
		var user model.User
		db.Where(&model.User{Id: lists[i].UserId}).First(&user)
		users = append(users, user)
	}

	/*
		    debug
			fmt.Printf("%+v\n", users)
			fmt.Printf("%+v\n", all_users)
	*/

	/*
	   all_usersはすべての人
	   usersは出席している人
	*/

	n := -1
	var absentUsers []model.User
	for i := 0; i < len(all_users); i++ {
		for j := 0; j < len(users); j++ {
			if all_users[i].Id == users[j].Id {
				n = j
			}
		}
		if n == -1 {
			absentUsers = append(absentUsers, all_users[i])
		}
		n = -1
	}

	log.Printf("ALLUSERLIST!!!!\n%+v", absentUsers)
	json.NewEncoder(w).Encode(absentUsers)
}
