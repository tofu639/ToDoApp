package model

import (
	"time"
)

// Todo represents a todo item in the system
type Todo struct {
	ID          uint      `json:"id" gorm:"primaryKey" example:"1"`
	Title       string    `json:"title" gorm:"not null;size:255" example:"Complete project"`
	Description string    `json:"description" gorm:"size:1000" example:"Finish the todo API backend project"`
	Completed   bool      `json:"completed" gorm:"default:false" example:"false"`
	UserID      uint      `json:"user_id" gorm:"not null;index" example:"1"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" example:"2024-01-01T12:00:00Z"`
	User        User      `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for the Todo model
func (Todo) TableName() string {
	return "todos"
}