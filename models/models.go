package models

import "time"

type User struct {
	ID           int       `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type SplitInput struct {
	UserID int     `json:"user_id"`
	Value  float64 `json:"value"`
}

type ExpenseRequest struct {
	Description string       `json:"description"`
	TotalAmount float64      `json:"total_amount"`
	PayerID     int          `json:"payer_id"`
	GroupID     *int         `json:"group_id,omitempty"`
	SplitType   string       `json:"split_type"`
	Splits      []SplitInput `json:"splits"`
}
