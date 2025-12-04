package usecases

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/usecases/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type FeedsUseCases interface {
	GetFeedUseCase
	NewFeedUseCase
	DeleteFeedUseCase
	HideFeedUseCase
	ShowFeedUseCase
}

type usecases struct {
	repo repo.FeedsRepo
}

func NewFeedsUsecase(repo repo.FeedsRepo) FeedsUseCases {
	return usecases{repo}
}

func (uc usecases) getFeedsExec(get func() (int, error)) (int, errro.Error) {
	n, err := get()
	if err == sql.ErrNoRows {
		e := errro.New(errro.ErrFeedsNoFeeds, "no feed(s) found")
		return 0, e
	}

	if err != nil {
		e := errro.FromError(errro.ErrFeedsDBError, "error fetching feeds", err)
		return 0, e
	}

	return n, nil
}

func (uc usecases) getFeedsByIDExec(ctx context.Context, id uuid.UUID, buf **entities.Feed) errro.Error {
	_, err := uc.getFeedsExec(func() (int, error) {
		err := uc.repo.GetFeedsByID(ctx, id, buf)
		if err != nil {
			return 0, err
		}
		return 1, nil
	})
	return err
}

func (uc usecases) getHidden(ctx context.Context, uname string, buf *[]entities.HiddenFeed) (int, errro.Error) {
	n, err := uc.repo.GetHidden(ctx, uname, buf)
	// if found eg. n != 0 && err == nil
	if err != nil {
		e := errro.FromError(errro.ErrFeedsGetHiddenFeeds, "failed to fetch hiddenfeeds", err)
		return n, e
	}

	return n, nil
}
