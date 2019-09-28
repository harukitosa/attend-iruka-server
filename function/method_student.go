package function

import (
	"Documents/attendance_book/server/src/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	//sqlite3をインポートします
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// _ "github.com/mattn/go-sqlite3"

// DatabaseName はデータベースの名前を保存します
var DatabaseName string

// DatabaseURL はデータベースのurlを保存します
var DatabaseURL string

// GenerateID uuidを生成して返す
func GenerateID() string {
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return "err"
	}
	uu := u.String()
	return uu
}

// DbInit データベース初期化する関数です
func DbInit() {
	//データベース関連
	// DatabaseURL = "test.sqlite3"
	// DatabaseName = "sqlite3"
	DatabaseURL = os.Getenv("DATABASE_URL")
	DatabaseName = "postgres"

	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（dbInit）")
	}
	//残りのモデルはまだ入れてない。
	db.AutoMigrate(&model.Student{})
	db.AutoMigrate(&model.Attendance{})
	defer db.Close()
}

// InsertStudent 一人の生徒情報を保存します
func InsertStudent(w http.ResponseWriter, r *http.Request) {
	var student model.Student
	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（InsertStudent）")
	}
	defer db.Close()
	log.Printf("POST: InsertStudent")

	vars := mux.Vars(r)
	decoder := json.NewDecoder(r.Body)

	error := decoder.Decode(&student)
	if error != nil {
		w.Write([]byte("json decode error" + error.Error() + "\n"))
	}
	student.OwnerID = vars["ownerID"]
	student.ID = GenerateID()
	db.Create(&model.Student{
		ID:            student.ID,
		Grade:         student.Grade,
		Class:         student.Class,
		Number:        student.Number,
		Name:          student.Name,
		OwnerID:       student.OwnerID,
		DefaultStatus: "attend",
	})
}

// GetStudents OwnerIDに対応した生徒情報を提供します
func GetStudents(w http.ResponseWriter, r *http.Request) {
	var students []model.Student
	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（GetStudents）")
	}
	defer db.Close()
	log.Printf("GET: GetStudents")

	vars := mux.Vars(r)
	ownerID := vars["ownerID"]

	db.Where(&model.Student{OwnerID: ownerID}).Find(&students)
	log.Printf("students:%+v", students)
	json.NewEncoder(w).Encode(students)
}

// RollCallAllStudents そのクラスの生徒の出席情報を記録します
func RollCallAllStudents(w http.ResponseWriter, r *http.Request) {

	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（RollCallAllStudents）")
	}
	defer db.Close()

	var student model.Student
	decoder := json.NewDecoder(r.Body)
	error := decoder.Decode(&student)
	if error != nil {
		w.Write([]byte("json decode error " + error.Error() + "\n"))
	}

	log.Printf("rollcall:%+v", student)

	uuid := GenerateID()
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	day := vars["day"]

	db.Create(&model.Attendance{
		ID:        uuid,
		StudentID: student.ID,
		Status:    student.DefaultStatus,
		Year:      year,
		Month:     month,
		Day:       day,
	})

}

// GetAttendanceMonthData すべての生徒の名前と指定された月の出席情報を提供します
func GetAttendanceMonthData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（GetAttendanceData）")
	}
	defer db.Close()

	vars := mux.Vars(r)
	ownerID := vars["ownerID"]
	year := vars["year"]
	month := vars["month"]

	// 生徒データを集める
	var students []model.Student
	db.Where(&model.Student{OwnerID: ownerID}).Find(&students)

	//指定された月の生徒の出席と欠席日数を求める
	var allAttendanceData []model.AllAttendanceData
	for i := 0; i < len(students); i++ {
		var countAttend int
		var countAbsent int
		var AttendanceList []model.Attendance
		var AbsentList []model.Attendance

		db.Where(&model.Attendance{StudentID: students[i].ID, Status: "attend", Year: year, Month: month}).Find(&AttendanceList)
		countAttend = len(AttendanceList)
		db.Where(&model.Attendance{StudentID: students[i].ID, Status: "absent", Year: year, Month: month}).Find(&AbsentList)
		countAbsent = len(AbsentList)
		var attendanceData model.AllAttendanceData
		attendanceData.StudentID = students[i].ID
		attendanceData.Name = students[i].Name
		attendanceData.Absent = strconv.Itoa(countAbsent)
		attendanceData.Attend = strconv.Itoa(countAttend)
		allAttendanceData = append(allAttendanceData, attendanceData)
	}
	log.Printf("all attendance data:%+v", allAttendanceData)
	json.NewEncoder(w).Encode(allAttendanceData)
}
