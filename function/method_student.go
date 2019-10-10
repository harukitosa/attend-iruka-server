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
	number := vars["number"]
	num, _ := strconv.Atoi(number)

	var nowStudents []model.Student
	db.Where(&model.Student{OwnerID: student.OwnerID}).Find(&nowStudents)

	for i := len(nowStudents); i < num+len(nowStudents); i++ {
		student.ID = GenerateID()
		n := strconv.Itoa(i + 1)
		db.Create(&model.Student{
			ID:            student.ID,
			Grade:         student.Grade,
			Class:         student.Class,
			Number:        n,
			OwnerID:       student.OwnerID,
			DefaultStatus: "attend",
		})
	}
}

// GetOneStudent 一人の生徒の出席データを返します
func GetOneStudent(w http.ResponseWriter, r *http.Request) {
	var attendance []model.Attendance

	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（GetOneStudent）")
	}
	defer db.Close()
	log.Printf("GET: GetOneStudents")

	vars := mux.Vars(r)
	id := vars["id"]

	db.Where(&model.Attendance{StudentID: id[1:]}).Find(&attendance)

	json.NewEncoder(w).Encode(attendance)
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

	var list model.Attendance
	db.Where(&model.Attendance{StudentID: student.ID, Year: year, Month: month, Day: day}).Find(&list)
	var check model.Check
	if list.Year == year {
		check.Check = false
		json.NewEncoder(w).Encode(check.Check)
	} else {
		db.Create(&model.Attendance{
			ID:        uuid,
			StudentID: student.ID,
			Status:    student.DefaultStatus,
			Year:      year,
			Month:     month,
			Day:       day,
		})
		check.Check = true
		json.NewEncoder(w).Encode(check.Check)
	}
}

// GetAttendanceRollData すべての生徒の名前と指定された月の出席情報を提供します
func GetAttendanceRollData(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（GetAttendanceData）")
	}
	defer db.Close()

	vars := mux.Vars(r)
	ownerID := vars["ownerID"]

	// 生徒データを集める
	var students []model.Student
	db.Where(&model.Student{OwnerID: ownerID}).Find(&students)

	//指定された月の生徒の出席と欠席日数を求める
	var allAttendanceData []model.AllAttendanceData
	for i := 0; i < len(students); i++ {
		var countAttend int
		var countAbsent int
		var attendanceList []model.Attendance
		var absentList []model.Attendance

		db.Where(&model.Attendance{StudentID: students[i].ID, Status: "attend"}).Find(&attendanceList)
		countAttend = len(attendanceList)
		db.Where(&model.Attendance{StudentID: students[i].ID, Status: "absent"}).Find(&absentList)
		countAbsent = len(absentList)
		var attendanceData model.AllAttendanceData
		attendanceData.StudentID = students[i].ID
		attendanceData.Name = students[i].Name
		attendanceData.Number = students[i].Number
		attendanceData.Absent = strconv.Itoa(countAbsent)
		attendanceData.Attend = strconv.Itoa(countAttend)
		allAttendanceData = append(allAttendanceData, attendanceData)
	}
	json.NewEncoder(w).Encode(allAttendanceData)
}

//UpdateAttendance this function update student date.
func UpdateAttendance(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open(DatabaseName, DatabaseURL)
	if err != nil {
		panic("We can't open database!（UpdateStudent）")
	}
	defer db.Close()
	log.Printf("POST: UpdateAttendance")
	var attendance model.Attendance
	decoder := json.NewDecoder(r.Body)
	error := decoder.Decode(&attendance)
	if error != nil {
		w.Write([]byte("json decode error " + error.Error() + "\n"))
	}
	log.Printf("$$:%v", attendance)
	// 条件付きでひとつのフィールドを更新します
	db.Model(&attendance).Where("ID = ?", attendance.ID).Update(model.Attendance{Status: attendance.Status})

}
