package usecases

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type HideFeedUseCase interface {
	// HideFeed put feed ID passed in [HideFeedRequest] into hiddenfeeds table.
	// If error occur returning error code either, [errro.ErrFeedsNoFeeds],
	// [errro.ErrFeedsMissingRequiredField] or [errro.ErrFeedsUnableToHide].
	HideFeed(ctx context.Context, hfr dto.HideFeedRequest) errro.Error
}

func (uc usecases) HideFeed(ctx context.Context, hfr dto.HideFeedRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "hideFeed-usecases")

	if hfr.FeedID.String() == "00000000-0000-0000-0000-000000000000" || hfr.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty ID")
		log.V(2).Info(e.Msg())
		return e
	}

	if hfr.Uname == "" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty Uname")
		log.V(2).Info(e.Msg())
		return e
	}

	// find feed with corresponding ID on the DB
	f := &entities.Feed{}
	err := uc.getFeedsByIDExec(ctx, hfr.FeedID, &f)
	if err != nil {
		e := errro.FromError(errro.ErrFeedsNoFeeds, "failed to fetch feed corresponding ID", err)
		return e
	}

	// hide it
	hf := &entities.HiddenFeed{ID: hfr.ID, FeedID: f.ID, To: hfr.Uname}
	if err := uc.repo.HideFeed(ctx, &hf); err != nil {
		e := errro.FromError(errro.ErrFeedsUnableToHide, "unable to hide feed", err)
		return e
	}

	return nil
}
