package recipes

import (
	"context"
	"sort"

	amqpAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/booleanism/tetek/comments/internal/repo"
	amqpFeeds "github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
)

type CommentsRecipes interface {
	GetComments(context.Context, GetCommentsRequest) ([]model.Comment, errro.Error)
	NewComment(context.Context, NewCommentRequest) (model.Comment, errro.Error)
}

type commRecipes struct {
	repo  repo.CommentsRepo
	feeds contracts.FeedsSubsribe
	auth  contracts.AuthSubscribe
}

func NewCommentRecipes(repo repo.CommentsRepo, feedsContr contracts.FeedsSubsribe, authContr contracts.AuthSubscribe) commRecipes {
	return commRecipes{repo, feedsContr, authContr}
}

func (cr commRecipes) GetComments(ctx context.Context, req GetCommentsRequest) ([]model.Comment, errro.Error) {
	ctx, log := loggr.GetLogger(ctx, "recipes")

	cf := repo.CommentFilter{
		Head:   req.Head,
		Offset: req.Offset,
	}

	var buf []model.Comment
	n, err := cr.repo.GetComments(ctx, cf, &buf)
	if err == nil && n != 0 {
		tree := buildCommentTree(buf, cf.Head)
		return tree, nil
	}

	if n <= 0 && err == nil {
		e := errro.New(errro.ECOMM_NO_COMM, "no such comments")
		log.V(1).Info(e.Error())
		return nil, e
	}

	// NOTE:
	// Handle known error here. we can utilizing errors.As
	// Also check other recipes with same pattern

	e := errro.FromError(errro.ECOMM_DB_ERR, "failed to fetch comments", err)
	return nil, e
}

func (cr commRecipes) NewComment(ctx context.Context, req NewCommentRequest) (model.Comment, errro.Error) {
	jwt := ctx.Value(keystore.AuthRes{}).(*amqpAuth.AuthResult)
	com := (newCommentRequest{
		NewCommentRequest: req,
		By:                jwt.Claims.Uname,
	}).toComment()

	// assume the head is feeds, then find it on feeds service
	t := amqpFeeds.FeedsTask{Cmd: 0, Feeds: amqpFeeds.Feeds{Id: req.Head}}
	feedsRes := &amqpFeeds.FeedsResult{Code: errro.EFEEDS_NO_FEEDS}
	err := cr.feedsAdapter(ctx, t, &feedsRes)
	if err != nil {
		return model.Comment{}, err
	}

	com.Parent = feedsRes.Detail.Id

	// if no feeds, we took into comments database it self
	if feedsRes.Code != errro.SUCCESS {
		var comBuf []model.Comment
		cf := repo.CommentFilter{Head: req.Head}
		if err := cr.getCommentsForNewComment(ctx, cf, &comBuf); err != nil {
			return model.Comment{}, err
		}

		// getCommentsForNewComment may returns
		// more than one comments, only last row of comments applied
		com.Parent = comBuf[len(comBuf)-1].Id
	}

	if err := cr.actualNewComment(ctx, &com, feedsRes); err != nil {
		return model.Comment{}, err
	}
	return com, nil
}

func (cr commRecipes) getCommentsForNewComment(ctx context.Context, cf repo.CommentFilter, comBuf *[]model.Comment) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "get-comment-for-new-comment")
	n, err := cr.repo.GetComments(ctx, cf, comBuf)
	if err != nil || n == 0 {
		e := errro.New(errro.ECOMM_NO_COMM, "there is no such comments")
		log.V(1).Info(e.Error(), "filter", cf)
		return e
	}
	return nil
}

func (cr commRecipes) actualNewComment(ctx context.Context, com *model.Comment, res *amqpFeeds.FeedsResult) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "actual-new-comment")
	com.Parent = res.Detail.Id
	err := cr.repo.NewComment(ctx, &com)
	if err != nil {
		e := errro.FromError(errro.ECOMM_DB_ERR, "failed to insert new comment", err)
		log.Error(err, e.Error(), "comment", com)
		return e
	}

	return nil
}

func buildCommentTree(tables []model.Comment, head uuid.UUID) []model.Comment {
	tree := []model.Comment{}
	for _, v := range tables {
		if v.Parent.String() == head.String() {
			v.Child = buildCommentTree(tables, v.Id)
			tree = append(tree, v)
		}
	}

	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Points > tree[j].Points
	})

	return tree
}
