package model

import (
	"github.com/google/uuid"
)

type CommentFilter struct {
	ID     uuid.UUID
	Head   uuid.UUID
	By     string
	Offset int
}

const DefaultLimit = 30
