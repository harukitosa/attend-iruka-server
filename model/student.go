package model

// Student 生徒情報
type Student struct {
	ID            string `json:"id" gorm:"PRIMARY_KEY"`
	Grade         string `json:"grade"`
	Class         string `json:"class"`
	Number        string `json:"number"`
	Name          string `json:"name"`
	OwnerID       string `json:"ownerID"`
	DefaultStatus string `json:"status"`
}

// List テスト用
type List struct {
	Student []Student `json:"students"`
}
