package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

	"Documents/attendance_book/server/src/function"
)

var CurrentUserId string

func main() {
	//port指定
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must beset")
	}
    
    function.DbInit()
	//CORS対応させるにはこの３つを加える必要がある。
    allowedOrigins := handlers.AllowedOrigins([]string{"https://sharp-wozniak-4e87de.netlify.com/register"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"})
	//Content-typeを加えるとpostできるようになる。
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type","X-Requested-with", "Authorization"})

	r := mux.NewRouter()
    //渡されるidの値に注意
	r.HandleFunc("/insert_user/:{id}", authMiddleware(function.InsertUserData)).Methods("POST", "OPTIONS")
	r.HandleFunc("/get_user/:{id}", authMiddleware(function.GetUserData)).Methods("GET")
	r.HandleFunc("/insert_room/:{id}", authMiddleware(function.InsertRoomData)).Methods("POST")
	r.HandleFunc("/get_owner_room/:{id}", authMiddleware(function.GetOwnerRoomData)).Methods("GET")
	r.HandleFunc("/insert_list/:{id}", authMiddleware(function.InsertListData)).Methods("POST")
    r.HandleFunc("/get_room_list/:{id}", authMiddleware(function.GetOwnerListData)).Methods("GET")
    r.HandleFunc("/check_room_pass/:{password}/:{userid}", authMiddleware(function.CheckRoomPassword)).Methods("GET")
    r.HandleFunc("/check_list_pass/:{password}/:{roomid}/:{userid}", authMiddleware(function.CheckListPassword)).Methods("GET")
    r.HandleFunc("/get_member_room/:{id}", function.GetMemberRoomData).Methods("GET")
    r.HandleFunc("/get_member_list/:{listid}", authMiddleware(function.GetListMember)).Methods("GET")
    r.HandleFunc("/get_all_member_list/:{listid}", authMiddleware(function.GetListAllMember)).Methods("GET")
	log.Printf("server start port localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)))
}

//firebaseで認証しているかどうかを確かめるミドルウェア
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Firebase SDK のセットアップ
		opt := option.WithCredentialsFile(os.Getenv("CREDENTIALS"))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Printf("error: %v\n", err)
			os.Exit(1)
		}
		auth, err := app.Auth(context.Background())
		if err != nil {
			log.Printf("error: %v\n", err)
			os.Exit(1)
		}
		// クライアントから送られてきた JWT 取得
		authHeader := r.Header.Get("Authorization")
		idToken := strings.Replace(authHeader, "Bearer ", "", 1)
		// JWT の検証
		token, err := auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			// JWT が無効なら Handler に進まず別処理
			fmt.Printf("error verifying ID token: %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("error verifying ID token\n"))
			return
		}
		// log.Printf("Verified ID token: %v\n", token)
		//user_idの受け取り方
		log.Printf("user id: %v\n", token.UID)
		CurrentUserId = token.UID
		next.ServeHTTP(w, r)
	}
}

