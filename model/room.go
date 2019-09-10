package model

type Room struct {
	Id          string `gorm:"PRIMARY_KEY" json:"id"`
	OwnerId     string `json:"owner_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Password    string `json:"password"`
}

type Check struct {
    Value bool `json:"value"`
}
