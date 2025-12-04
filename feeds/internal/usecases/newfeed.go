package usecases

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type NewFeedUseCase interface {
	// Put feed into feeds by transforming [NewFeedRequest] into [entities.Feeds].
	// If error occur returning error code either [errro.ErrFeedsNewFail],
	// [errro.ErrFeedsMissingRequiredField] or [errro.ErrFeedsInvalidType].
	NewFeed(ctx context.Context, nfr dto.NewFeedRequest) errro.Error
}

// It's checks for nfr.ID, nfr.Uname and nfr.Type
func (uc usecases) NewFeed(ctx context.Context, nfr dto.NewFeedRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "newFeeds-usecases")

	if nfr.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty ID")
		log.V(2).Info(e.Msg())
		return e
	}

	if nfr.Uname == "" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty Uname")
		log.V(2).Info(e.Msg())
		return e
	}

	f := nfr.ToFeed()
	isValidType := false
	for _, v := range model.AvailableType {
		isValidType = (f.Type == v) || isValidType
	}

	if !isValidType {
		e := errro.New(errro.ErrFeedsInvalidType, "invalid type, type should either M, J, S, A")
		log.V(2).Info(e.Msg())
		return e
	}

	fp := &f
	err := uc.repo.PutIntoFeed(ctx, &fp)
	if err != nil {
		e := errro.New(errro.ErrFeedsNewFail, "unable to insert new feed")
		return e
	}

	return nil
}
