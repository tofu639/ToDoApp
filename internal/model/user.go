package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey" example:"1"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null;size:255" example:"user@example.com"`
	Password  string    `json:"-" gorm:"not null;size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" example:"2024-01-01T12:00:00Z"`
	Todos     []Todo    `json:"todos,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// UserInfo represents user information for API responses (without password)
type UserInfo struct {
	ID        uint      `json:"id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ToUserInfo converts a User to UserInfo (removes sensitive data)
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}