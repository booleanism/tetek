package usecases

import (
	"context"
	"time"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type DeleteFeedUseCase interface {
	// DeleteFeed change feeds' deleted_at state into [DeleteFeedRequest]'s DeletedAt.
	// If error occur returning error code either [errro.ErrFeedsUnathorized],
	// [errro.ErrFeedsMissingRequiredField], [errro.ErrFeedsNoFeeds] or [errro.ErrFeedsDeleteFail].
	DeleteFeed(ctx context.Context, dfr dto.DeleteFeedRequest) errro.Error
}

func (uc usecases) DeleteFeed(ctx context.Context, dfr dto.DeleteFeedRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "deleteFeed-usecases")

	if dfr.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty ID")
		log.V(2).Info(e.Msg())
		return e
	}

	f := &entities.Feed{}
	err := uc.getFeedsByIDExec(ctx, dfr.ID, &f)
	if err != nil {
		e := errro.FromError(errro.ErrFeedsNoFeeds, "failed to fetch feed by ID", err)
		return e
	}

	if f.By != dfr.Uname && dfr.Role != "M" {
		e := errro.New(errro.ErrFeedsUnathorized, "unauthorized to delete feed")
		log.V(2).Info(e.Msg(), "feed", f, "request", dfr)
		return e
	}

	fd := model.FeedDeletion{ID: dfr.ID, DeletedAt: time.Now()}
	if err := uc.repo.DeleteFeed(ctx, fd, &f); err != nil {
		e := errro.FromError(errro.ErrFeedsDeleteFail, "failed to delete feed", err)
		return e
	}

	return nil
}
