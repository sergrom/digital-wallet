package main

import "time"

// User ...
type User struct {
	UserID    int64     `db:"user_id"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

// CreateUserReq ...
type CreateUserReq struct {
	Email string `json:"email"`
}

// UserCreatedMsg ...
type UserCreatedMsg struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
