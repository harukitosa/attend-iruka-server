# IRUKA System

## サーバーサイド編

### main.go

url

```go
/insert_user/:{id}
/get_user/:{id}
/insert_room/:{id}
/get_owner_room/:{id}
/insert_list/:{id}
/get_room_list/:{id}
/check_room_pass/:{password}/:{userid}
/check_list_pass/:{password}/:{roomid}/:{useid}
/get_member_room/:{id}
/get_member_list/:{listid}
```

### package functon

method_list.go

```
1
InsertListData(w http.ResponseWriter, r *http.Request)
POST db.Create(List) db.Create(PasswordList)
//listの新規作成

2.
GetOwnerListData(w http.ResponseWriter, r *http.Request)
GET Json model.List
//ユーザーが所有しているroomの取り出し

3.
CheckListPassword(w http.ResponseWriter, r *http.Request)
//listのパスワード照合とRelationUserListの作成
```
