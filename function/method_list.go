package function

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
		Id:       list.Id,
		RoomId:   list.RoomId,
		Title:    list.Title,
	})
	db.Create(&model.PasswordList{
        ListId: list.Id,
        UserId: password.UserId,
        Password: password.Password,
        Time:   t,
    })
    var all []model.PasswordList
    db.Find(&all)
    fmt.Printf("all pass list:%+v\n", all)
	fmt.Printf("list:%+v\n", list)
    fmt.Printf("passlist:%+v\n", password)
}

//所有しているルームのデータ取り出し
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

//listのパスワードを取り出す
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

    //debug
    log.Println("password:"+password)
    log.Println("room_id:"+room_id)
    log.Println("user_id:"+user_id)
    var check model.Check

    var lists []model.List
    db.Where(&model.List{RoomId: room_id}).Find(&lists)
    
    var passwordLists []model.PasswordList
    for i:=0;i<len(lists);i++ {
        var passlist model.PasswordList
        db.Where(&model.PasswordList{ListId: lists[i].Id}).First(&passlist)
        fmt.Printf("pass:%+v\n", passlist)
        if(passlist.Password == password) {
            passwordLists = append(passwordLists, passlist)      
        }
    }

    //debug
    for i:=0;i<len(passwordLists);i++ {
        fmt.Printf("passwordList:%+v\n", passwordLists[i])
    }
    
    if len(passwordLists) == 0 {
        check.Value = false
        json.NewEncoder(w).Encode(check)
        return
    } 
    now := time.Now()
    back := passwordLists[0].Time
    timepass := back.Add(15 *time.Minute)
    
    if now.Before(timepass) {
        //時間が間に合っているとき
        //jsonでtrueをかえす
        check.Value = InsertRelationUserList(user_id, passwordLists[0].ListId)
        log.Println("time is ok")
        json.NewEncoder(w).Encode(check)
        return
        
    } else {
        //時間が間に合っていないとき
        //jsonでfalseを返す
        check.Value = false
        log.Println("time is bas")
        json.NewEncoder(w).Encode(check)
        return
    }
}


func InsertRelationUserList(user_id string, list_id string) bool{
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
    db.Where(&model.RelationUserList{UserId: user_id, ListId:list_id}).Find(&checks)
    for i:=0;i < len(checks);i++ {
        fmt.Printf("checks:%+v\n", checks[i])
    }
    //nilのときはokそれ以外はuser_idが等しいことを利用して弾く
    for i:=0;i < len(checks);i++ {
        if checks[i].UserId == user_id {
            return false
        }
    }

    data := model.RelationUserList{
        Id: uu,
        UserId: user_id,
        ListId: list_id,
    }
    fmt.Printf("data:%+v\n", data)
    db.Create(&data)
    return true
}

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
    db.Where(&model.RelationUserList{ListId:list_id}).Find(&lists)
    for i:=0;i < len(lists);i++ {
        fmt.Printf("checks:%+v\n", lists[i])
    }

    var users []model.User
    for i:=0;i < len(lists);i++ {
        var user model.User
        db.Where(&model.User{Id: lists[i].UserId}).First(&user)
        users = append(users, user)
    }
    
    json.NewEncoder(w).Encode(users)
}

func GetListAllMember(w http.ResponseWriter, r *http.Request) {
    //データベースを開く
	db, err := gorm.Open(DatabaseName, DatabaseUrl)
	if err != nil {
		panic("We can't open database GetListAllMember")
	}
	defer db.Close()
    log.Printf("GET: GetListAllMember")
    vars := mux.Vars(r)
    list_id := vars["listid"]
    
    //以下ユーザー情報の全取得
    var num model.List
    db.Where(&model.List{Id: list_id}).First(&num)
    var list_room model.Room
    db.Where(&model.Room{Id: num.RoomId}).First(&list_room)
    var all_users_id []model.RelationUserRoom
    db.Where(&model.RelationUserRoom{RoomId: list_room.Id}).Find(&all_users_id)
    log.Printf("#####################\n%+v", all_users_id)
    
    var all_users []model.User
    for i:=0;i<len(all_users_id);i++ {
        var all_user model.User
        db.Where(&model.User{Id: all_users_id[i].UserId}).First(&all_user)
        all_users = append(all_users, all_user)
    }

    var lists []model.RelationUserList
    db.Where(&model.RelationUserList{ListId:list_id}).Find(&lists)
    for i:=0;i < len(lists);i++ {
        fmt.Printf("checks:%+v\n", lists[i])
    }

    var users []model.User
    for i:=0;i < len(lists);i++ {
        var user model.User
        db.Where(&model.User{Id: lists[i].UserId}).First(&user)
        users = append(users, user)
    }

    fmt.Printf("%+v\n", users)
    fmt.Printf("%+v\n", all_users)
    
    n := -1
    var absent_users []model.User
    for i:=0; i<len(all_users); i++ {
        for j:=0; j<len(users); j++ {
            if(all_users[i].Id == users[j].Id) {
                n = j
            }
        }
        if(n == -1) {
            absent_users = append(absent_users, all_users[i])
        }
        n = -1
    }
    
    
    log.Printf("ALLUSERLIST!!!!\n%+v", absent_users)
    json.NewEncoder(w).Encode(absent_users)
}

