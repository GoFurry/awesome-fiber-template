package models

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Email     string    `gorm:"size:128;not null;uniqueIndex" json:"email"`
	Age       int       `gorm:"not null;default:18" json:"age"`
	Status    string    `gorm:"size:32;not null;default:active" json:"status"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "demo_users"
}
