package main

import (
	"context"
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

// CurrentUserID 現在ログインしているユーザーのIDを保存します
var CurrentUserID string

func main() {
	//port指定
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must beset")
	}

	function.DbInit()
	//CORS対応させるにはこの３つを加える必要がある。
	//Content-typeを加えるとpostできるようになる。
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:8080", "https://iruka-roll-book.com"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Origin", "Content-Type", "X-Requested-with", "Authorization"})
	r := mux.NewRouter()

	r.HandleFunc("/insert_student/:{ownerID}/:{number}", authMiddleware(function.InsertStudent)).Methods("POST")
	r.HandleFunc("/get_students/:{ownerID}", authMiddleware(function.GetStudents)).Methods("GET")
	r.HandleFunc("/roll_call/:{year}/:{month}/:{day}", authMiddleware(function.RollCallAllStudents)).Methods("POST")
	r.HandleFunc("/get_roll_data/:{ownerID}", authMiddleware(function.GetAttendanceRollData)).Methods("GET")

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
			log.Printf("error verifying ID token: %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("error verifying ID token\n"))
			return
		}
		// log.Printf("Verified ID token: %v\n", token)
		//user_idの受け取り方
		// log.Printf("user id: %v\n", token.UID)
		CurrentUserID = token.UID
		next.ServeHTTP(w, r)
	}
}
