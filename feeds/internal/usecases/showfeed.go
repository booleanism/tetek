package usecases

import (
	"context"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/feeds/internal/domain/model/pools"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
)

type ShowFeedUseCase interface {
	// ShowFeed only show fees if [ShowFeedRequest] not in hiddenfeeds for spesific uname.
	// If error occur returning error code either:
	// - [errro.ErrFeedsMissingRequiredField]
	// - [errro.ErrAcqPool]
	// - [errro.ErrFeedsGetHiddenFeeds]
	// - [errro.ErrFeedsHidden]
	// - [errro.ErrAccountServiceUnavailable]
	// - [errro.ErrFeedsNoFeeds]
	// - [errro.ErrFeedsDBError]
	ShowFeed(ctx context.Context, commsDealer contracts.CommentsDealer, commAdapter adapter.CommentsAdapterFn, sfr dto.ShowFeedRequest, buf **model.FeedWithComments) errro.Error
}

func (uc usecases) ShowFeed(ctx context.Context, commsDealer contracts.CommentsDealer, commAdapter adapter.CommentsAdapterFn, sfr dto.ShowFeedRequest, buf **model.FeedWithComments) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "hideFeed-usecases")

	if sfr.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty ID")
		log.V(2).Info(e.Msg())
		return e
	}

	if sfr.Uname == "" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "empty ID")
		log.V(2).Info(e.Msg())
		return e
	}

	hfBuf, ok := pools.HiddenFeedsPool.Get().(*pools.HiddenFeeds)
	if !ok {
		e := errro.New(errro.ErrAcqPool, "failed to acquire hiddenfeeds pool")
		log.Error(e, e.Msg(), "request", sfr)
		return e
	}
	defer pools.HiddenFeedsPool.Put(hfBuf)
	defer hfBuf.Reset()

	n, err := uc.getHidden(ctx, sfr.Uname, &hfBuf.Value)
	if err != nil {
		log.V(2).Info("unable to fetch visible feed", "n", n, "error", err)
		return err
	}

	log.Info("it's hidden?", "hf", hfBuf.Value[:n], "id", sfr.ID)
	// Does this feeds not hidden?
	feedsHidden := false
	for _, v := range hfBuf.Value[:n] {
		feedsHidden = (v.ID == sfr.ID) || feedsHidden
	}

	// If yes then return error
	if feedsHidden {
		e := errro.New(errro.ErrFeedsHidden, "feed hidden")
		log.V(2).Info(e.Msg(), "n", n)
		return e
	}

	fBuf := &(*buf).Feed
	if err := uc.getFeedsByIDExec(ctx, sfr.ID, &fBuf); err != nil {
		return err
	}

	cBuf := &amqp.CommentsResult{}
	task := amqp.CommentsTask{}
	if err := commAdapter(ctx, commsDealer, task, &cBuf); err != nil {
		return err
	}

	comms := helper.BuildCommentTree(cBuf.Details, fBuf.ID)
	(*buf).Comments = comms

	return nil
}
