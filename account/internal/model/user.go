package model

import "time"

type User struct {
	ID        string     `json:"id"`
	Uname     string     `json:"uname"`
	Email     string     `json:"email"`
	Passwd    string     `json:"passwd"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"created_at"`
}
