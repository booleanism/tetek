package model

import (
	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/feeds/internal/entities"
)

type FeedWithComments struct {
	entities.Feed
	Comments []amqp.Comment `json:"comments"`
}
