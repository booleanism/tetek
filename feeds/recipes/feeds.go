package recipes

import (
	"context"
	"strings"

	mqAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/pools"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FeedRecipes interface {
	Feeds(context.Context, GetFeedsRequest, *pools.Feeds) errro.Error
	New(context.Context, NewFeedRequest) errro.Error
	Delete(context.Context, DeleteRequest) errro.Error
	Hide(context.Context, HideRequest) errro.Error
	Shows(context.Context, ShowFeedRequest, **model.FeedWithComments) errro.Error
}

type feedRecipes struct {
	repo      repo.FeedsRepo
	commContr contracts.CommentsSubscribe
}

func NewRecipes(repo repo.FeedsRepo, commContr contracts.CommentsSubscribe) feedRecipes {
	return feedRecipes{repo, commContr}
}

func (fr feedRecipes) Feeds(ctx context.Context, req GetFeedsRequest, f *pools.Feeds) errro.Error {
	ctx, _ = loggr.GetLogger(ctx, "getFeeds-recipes")
	ff := repo.FeedsFilter{
		Offset: uint64(req.Offset),
		Type:   req.Type,
		ID:     req.ID,
	}

	jwt, ok := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	if ok {
		ff.HiddenTo = jwt.Claims.Uname
	}

	if ff.Type == "" {
		ff.Type = "M"
	}
	ff.Type = strings.ToUpper(ff.Type)

	return fr.getFeeds(ctx, ff, &f.Value)
}

func (fr feedRecipes) getFeeds(ctx context.Context, ff repo.FeedsFilter, f *[]model.Feed) errro.Error {
	err := fr.repo.Feeds(ctx, ff, f)
	if err == nil && len(*f) != 0 {
		return nil
	}

	if err != pgx.ErrNoRows {
		e := errro.FromError(errro.ErrFeedsDBError, "error fetching feeds", err)
		return e
	}

	if len(*f) == 0 || err == pgx.ErrNoRows {
		e := errro.New(errro.ErrFeedsNoFeeds, "no feed(s) found")
		return e
	}

	return nil
}

func (fr feedRecipes) New(ctx context.Context, req NewFeedRequest) errro.Error {
	ctx, _ = loggr.GetLogger(ctx, "newFeed-recipes")

	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)

	rFeed := req.ToFeed()
	rFeed.By = jwt.Claims.Uname
	rFeed.Type = strings.ToUpper(rFeed.Type)
	f := &rFeed

	err := fr.repo.NewFeed(ctx, &f)
	if err != nil {
		e := errro.FromError(errro.ErrFeedsDBError, "could not insert new feed", err)
		return e
	}

	return nil
}

func (fr feedRecipes) Delete(ctx context.Context, req DeleteRequest) errro.Error {
	ff := repo.FeedsFilter{
		ID: req.ID,
	}

	if ff.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "missing required field")
		return e
	}

	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	// only moderator freely to delete feed
	if strings.ToLower(jwt.Claims.Role) != "m" {
		ff.By = jwt.Claims.Uname
	}

	fBuf, ok := pools.FeedsPool.Get().(*pools.Feeds)
	if !ok {
		e := errro.New(errro.ErrAcqPool, "failed to acquire pool")
		return e
	}
	defer pools.FeedsPool.Put(fBuf)
	defer fBuf.Reset()

	f := &(*fBuf).Value[0]
	err := fr.repo.DeleteFeed(ctx, ff, &f)
	if err == nil {
		return nil
	}

	if err == pgx.ErrNoRows {
		e := errro.New(errro.ErrFeedsNoFeeds, "failed to delete feed: no row")
		return e
	}

	if f.DeletedAt == nil {
		e := errro.New(errro.ErrFeedsNoFeeds, "no such feed")
		return e
	}

	e := errro.New(errro.ErrFeedsDBError, "somthing happen when trying to delete feed")
	return e
}

// TODO:
// Validate feed by req.Id
func (fr feedRecipes) Hide(ctx context.Context, req HideRequest) errro.Error {
	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	hfBuf, ok := pools.HiddenFeedsPool.Get().(*pools.HiddenFeeds)
	if !ok {
		e := errro.New(errro.ErrAcqPool, "failed to acquire pool")
		return e
	}
	defer pools.HiddenFeedsPool.Put(hfBuf)
	defer hfBuf.Reset()

	hf := &hfBuf.Value[0]
	hf.ID = uuid.New()
	hf.To = jwt.Claims.Uname
	hf.FeedID = req.ID

	er := fr.repo.HideFeed(ctx, &hf)
	if er != nil {
		e := errro.FromError(errro.ErrFeedsDBError, "unable to hide feed", er)
		return e
	}

	return nil
}

func (fr feedRecipes) Shows(ctx context.Context, req ShowFeedRequest, f **model.FeedWithComments) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "showsFeed-recipes")
	t := amqp.CommentsTask{
		Cmd: 0,
		Comment: amqp.Comment{
			Parent: req.ID,
		},
	}

	fBuf, ok := pools.FeedsPool.Get().(*pools.Feeds)
	if !ok {
		e := errro.New(errro.ErrAcqPool, "failed to acquire pool")
		log.Error(e, e.Msg(), "task", t)
		return e
	}

	defer pools.FeedsPool.Put(fBuf)
	defer fBuf.Reset()

	ff := repo.FeedsFilter{ID: req.ID}

	if err := fr.getFeeds(ctx, ff, &fBuf.Value); err != nil {
		return err
	}

	res := &amqp.CommentsResult{}
	if err := fr.commAdapter(ctx, t, &res); err != nil {
		return err
	}

	comms := helper.BuildCommentTree(res.Details, fBuf.Value[0].ID)

	(*f).Feed = fBuf.Value[0]
	(*f).Comments = comms

	return nil
}
