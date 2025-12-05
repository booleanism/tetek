package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id,omitempty"`
	Uname     string     `json:"uname,omitempty"`
	Email     string     `json:"email,omitempty"`
	Passwd    string     `json:"passwd,omitempty"`
	Role      string     `json:"role,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
