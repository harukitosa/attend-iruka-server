package model

// Attendance 出欠状況
type Attendance struct {
	ID        string `json:"id" gorm:"PRIMARY_KEY"`
	StudentID string `json:"studentID"`
	Status    string `json:"status"`
	Year      string `json:"year"`
	Month     string `json:"month"`
	Day       string `json:"day"`
}

// AllAttendanceData json形式で送りやすいようにデータをまとめます
type AllAttendanceData struct {
	StudentID string `json:"studentID"`
	Name      string `json:"name"`
	Attend    string `json:"attend"`
	Absent    string `json:"absent"`
}
