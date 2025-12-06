package usecases

import (
	"context"
	"time"

	"github.com/booleanism/tetek/comments/internal/internal/domain/model"
	"github.com/booleanism/tetek/comments/internal/usecases/dto"
	messaging "github.com/booleanism/tetek/feeds/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type NewCommentUseCase interface {
	NewComment(ctx context.Context, feedsDealer contracts.FeedsDealer, feedsAdapter adapter.FeedsAdapterFn, rc dto.NewCommentRequest) errro.Error
}

func (uc usecases) NewComment(ctx context.Context, feedsDealer contracts.FeedsDealer, feedsAdapter adapter.FeedsAdapterFn, nrc dto.NewCommentRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "putComment-usecases")

	if nrc.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrCommMissingRequiredField, "invalid comment ID")
		log.V(2).Info(e.Msg())
		return e
	}

	if nrc.Head.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrCommMissingRequiredField, "invalid Head ID")
		log.V(2).Info(e.Msg())
		return e
	}

	if nrc.Text == "" || nrc.By == "" {
		e := errro.New(errro.ErrCommMissingRequiredField, "missing required field")
		log.V(2).Info(e.Msg())
		return e
	}

	// assume the head is feeds, then find it on feeds service
	t := messaging.FeedsTask{Cmd: 0, Feeds: messaging.Feeds{ID: nrc.Head}}
	feedsRes := &messaging.FeedsResult{Code: errro.ErrFeedsNoFeeds}
	err := feedsAdapter(ctx, feedsDealer, t, &feedsRes)
	if err != nil {
		return err
	}

	// if no feeds, we took into comments database it self
	if feedsRes.Code != errro.Success {
		var comBuf []dto.Comment
		cf := dto.CommentFilter{Head: nrc.Head}
		if err := uc.getCommentsHead(ctx, cf, &comBuf); err != nil {
			return err
		}

		// getCommentsForNewComment may returns
		// more than one comments, only last row of comments applied
		nrc.Head = comBuf[len(comBuf)-1].ID
	} else {
		nrc.Head = feedsRes.Details[len(feedsRes.Details)-1].ID
	}

	now := time.Now()
	buf := &dto.Comment{
		ID:        nrc.ID,
		Parent:    nrc.Head,
		By:        nrc.By,
		Text:      nrc.Text,
		CreatedAt: &now,
	}

	log.Info("buf", "buf", buf)
	if err := uc.repo.PutComment(ctx, &buf); err != nil {
		e := errro.FromError(errro.ErrCommNewFail, "failed to put comment", err)
		return e
	}

	return nil
}

func (uc usecases) getCommentsHead(ctx context.Context, cf dto.CommentFilter, comBuf *[]dto.Comment) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "getCommentsHead")
	n, err := uc.repo.GetComments(ctx, cf.ID, model.DefaultLimit, cf.Offset, comBuf)
	if err != nil || n == 0 {
		e := errro.New(errro.ErrCommNoComments, "there is no such comments")
		log.V(2).Info(e.Error(), "filter", cf)
		return e
	}
	return nil
}
