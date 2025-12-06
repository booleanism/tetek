package model

import (
	"time"

	messaging "github.com/booleanism/tetek/comments/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/google/uuid"
)

const DefaultLimit = 30

var AvailableType = []string{"M", "J", "S", "A"}

type FeedWithComments struct {
	entities.Feed
	Comments []messaging.Comment `json:"comments"`
}

type FeedDeletion struct {
	ID        uuid.UUID `json:"id,omitempty"`
	DeletedAt time.Time `json:"deleted_at"`
}

type GetFeedsByType struct {
	Type   string
	Limit  int
	Offset int
}
