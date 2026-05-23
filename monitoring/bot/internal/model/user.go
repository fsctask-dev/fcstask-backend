package user

import "time"

type User struct {
	ChatID    int64     `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}
