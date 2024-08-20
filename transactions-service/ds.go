package main

import "time"

// User ...
type User struct {
	UserID    int64     `db:"user_id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
}

// UserCreatedMsg ...
type UserCreatedMsg struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AddMoneyReq ...
type AddMoneyReq struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

// TransferMoneyReq ...
type TransferMoneyReq struct {
	FromUserID       int64   `json:"from_user_id"`
	ToUserID         int64   `json:"to_user_id"`
	AmountToTransfer float64 `json:"amount_to_transfer"`
}
