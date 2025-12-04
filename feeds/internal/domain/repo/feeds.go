package repo

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/google/uuid"
)

type FeedsEntityRepo interface {
	FeedsGetter
	FeedAdder
	FeedsUpdater
}

type FeedsGetter interface {
	GetFeedsByID(ctx context.Context, id uuid.UUID, buf **entities.Feed) error

	// GetFeedsByType return n scanned rows or an error.
	// Scanned rows placed in buf and it's should not nil.
	// [model.GetFeedsByType] property should not filled with default/empty value.
	GetFeedsByType(ctx context.Context, fType model.GetFeedsByType, buf *[]entities.Feed) (n int, err error)

	// GetFeedsNotHiddenBy return n scanned rows or an error.
	// It's similar to GetFeedsByType but with string param which is represent username.
	GetFeedsNotHiddenBy(ctx context.Context, fType model.GetFeedsByType, by string, buf *[]entities.Feed) (n int, err error)
}

type FeedAdder interface {
	PutIntoFeed(ctx context.Context, feed **entities.Feed) error
}

type FeedsUpdater interface {
	DeleteFeed(ctx context.Context, fd model.FeedDeletion, feed **entities.Feed) error
}
