package model

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type ProjectIdea struct {
	ID          int64
	UserID      int64
	UserEmail   string
	Title       string
	Description string
	CreatedAt   time.Time
}
