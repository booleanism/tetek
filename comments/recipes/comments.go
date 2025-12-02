package recipes

import (
	"context"

	amqpAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/booleanism/tetek/comments/internal/repo"
	amqpFeeds "github.com/booleanism/tetek/feeds/infra/amqp"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
)

type CommentsRecipes interface {
	NewComment(context.Context, NewCommentRequest) (model.Comment, errro.Error)
}

type commRecipes struct {
	repo  repo.CommentsRepo
	feeds contracts.FeedsDealer
	auth  contracts.AuthDealer
}

func NewCommentRecipes(repo repo.CommentsRepo, feedsContr contracts.FeedsDealer, authContr contracts.AuthDealer) commRecipes {
	return commRecipes{repo, feedsContr, authContr}
}

func (cr commRecipes) NewComment(ctx context.Context, req NewCommentRequest) (model.Comment, errro.Error) {
	jwt := ctx.Value(keystore.AuthRes{}).(*amqpAuth.AuthResult)
	com := (newCommentRequest{
		NewCommentRequest: req,
		By:                jwt.Claims.Uname,
	}).toComment()

	// assume the head is feeds, then find it on feeds service
	t := amqpFeeds.FeedsTask{Cmd: 0, Feeds: amqpFeeds.Feeds{ID: req.Head}}
	feedsRes := &amqpFeeds.FeedsResult{Code: errro.ErrFeedsNoFeeds}
	err := adapter.FeedsAdapter(ctx, cr.feeds, t, &feedsRes)
	if err != nil {
		return model.Comment{}, err
	}

	com.Parent = feedsRes.Detail.ID

	// if no feeds, we took into comments database it self
	if feedsRes.Code != errro.Success {
		var comBuf []model.Comment
		cf := repo.CommentFilter{Head: req.Head}
		if err := cr.getCommentsHead(ctx, cf, &comBuf); err != nil {
			return model.Comment{}, err
		}

		// getCommentsForNewComment may returns
		// more than one comments, only last row of comments applied
		com.Parent = comBuf[len(comBuf)-1].ID
	}

	if err := cr.actualNewComment(ctx, &com, feedsRes); err != nil {
		return model.Comment{}, err
	}
	return com, nil
}

func (cr commRecipes) getCommentsHead(ctx context.Context, cf repo.CommentFilter, comBuf *[]model.Comment) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "get-comment-for-new-comment")
	n, err := cr.repo.GetComments(ctx, cf, comBuf)
	if err != nil || n == 0 {
		e := errro.New(errro.ErrCommNoConsume, "there is no such comments")
		log.V(1).Info(e.Error(), "filter", cf)
		return e
	}
	return nil
}

func (cr commRecipes) actualNewComment(ctx context.Context, com *model.Comment, res *amqpFeeds.FeedsResult) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "actual-new-comment")
	com.Parent = res.Detail.ID
	err := cr.repo.NewComment(ctx, &com)
	if err != nil {
		e := errro.FromError(errro.ErrCommDBError, "failed to insert new comment", err)
		log.Error(err, e.Error(), "comment", com)
		return e
	}

	return nil
}
