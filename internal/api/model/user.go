package model

import (
	"time"
)

type User struct {
	ID        int       `json:"-" db:"id"`
	Name      string    `json:"name" binding:"required"`
	Username  string    `json:"username" binding:"required"` // for QR
	Password  string    `json:"password" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}
