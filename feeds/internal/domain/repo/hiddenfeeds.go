package repo

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
)

type HiddenFeedsGetter interface {
	GetHidden(ctx context.Context, by string, buf *[]entities.HiddenFeed) (n int, err error)
}

type HiddenFeedsEntityRepo interface {
	HiddenFeedsGetter
	FeedsHidder
}

type FeedsHidder interface {
	HideFeed(ctx context.Context, hf **entities.HiddenFeed) error
}
