package recipes

import (
	"context"
	"errors"
	"strings"

	"github.com/booleanism/tetek/account/amqp"
	mqAcc "github.com/booleanism/tetek/account/amqp"
	mqAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type FeedRecipes interface {
	Feeds(context.Context, GetFeedsRequest) ([]model.Feed, errro.Error)
	New(context.Context, NewFeedRequest) errro.Error
	Delete(context.Context, DeleteRequest) errro.Error
	Hide(context.Context, HideRequest) errro.Error
}

type feedRecipes struct {
	repo     repo.FeedsRepo
	accContr contracts.AccountSubscribe
}

func NewRecipes(repo repo.FeedsRepo, accContr contracts.AccountSubscribe) *feedRecipes {
	return &feedRecipes{repo, accContr}
}

func (fr *feedRecipes) Feeds(ctx context.Context, req GetFeedsRequest) ([]model.Feed, errro.Error) {
	ff := repo.FeedsFilter{
		Offset: uint64(req.Offset),
		Type:   req.Type,
		Id:     req.Id,
	}

	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	var res *mqAcc.AccountRes
	if jwt != nil {
		t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
		err := fr.accAdapter(ctx, t, &res)
		if err != nil {
			return nil, err
		}
	}

	if ff.Type == "" {
		ff.Type = "M"
	}
	ff.Type = strings.ToUpper(ff.Type)
	if res != nil {
		ff.HiddenTo = res.Detail.Uname
	}

	f, er := fr.repo.Feeds(ctx, ff)
	if er != nil {
		var pgErr *pgconn.PgError
		if !errors.As(er, &pgErr) {
			e := errro.FromError(errro.EFEEDS_DB_ERR, "error fetching feeds", er)
			return nil, e
		}

		if pgErr.Code == "23505" {
			e := errro.New(errro.EFEEDS_NO_FEEDS, "no feed(s) found")
			return nil, e
		}
	}

	if len(f) == 0 {
		e := errro.New(errro.EFEEDS_NO_FEEDS, "no feed(s) found")
		return nil, e
	}

	return f, nil
}

func (fr *feedRecipes) New(ctx context.Context, req NewFeedRequest) errro.Error {
	rFeed := req.ToFeed()
	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	var res *amqp.AccountRes
	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	err := fr.accAdapter(ctx, t, &res)
	if err != nil {
		return err
	}

	rFeed.By = res.Detail.Uname
	rFeed.Type = strings.ToUpper(rFeed.Type)

	_, er := fr.repo.NewFeed(ctx, rFeed)
	if er != nil {
		e := errro.FromError(errro.EFEEDS_DB_ERR, "could not insert new feed", er)
		return e
	}

	return nil
}

func (fr *feedRecipes) Delete(ctx context.Context, req DeleteRequest) errro.Error {
	ff := repo.FeedsFilter{
		Id: req.Id,
	}

	if ff.Id.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.EFEEDS_MISSING_REQUIRED_FIELD, "missing required field")
		return e
	}

	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	var res *amqp.AccountRes
	err := fr.accAdapter(ctx, t, &res)
	if err != nil {
		return err
	}

	// only moderator freely to delete feed
	if strings.ToLower(res.Detail.Uname) != "m" {
		ff.By = jwt.Claims.Uname
	}

	fDel, er := fr.repo.DeleteFeed(ctx, ff)
	if er == nil {
		return nil
	}

	if er == pgx.ErrNoRows {
		e := errro.New(errro.EFEEDS_NO_FEEDS, "failed to delete feed: no row")
		return e
	}

	if fDel.Deleted_At == nil {
		e := errro.New(errro.EFEEDS_NO_FEEDS, "no such feed")
		return e
	}

	e := errro.New(errro.EFEEDS_DB_ERR, "somthing happen when trying to delete feed")
	return e
}

// TODO:
// Validate feed by req.Id
func (fr *feedRecipes) Hide(ctx context.Context, req HideRequest) errro.Error {
	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	var res *mqAcc.AccountRes
	err := fr.accAdapter(ctx, t, &res)
	if err != nil {
		return err
	}

	hf := repo.HiddenFeeds{
		Id:     uuid.NewString(),
		To:     res.Detail.Uname,
		FeedId: req.Id,
	}

	_, er := fr.repo.HideFeed(ctx, hf)
	if er != nil {
		e := errro.FromError(errro.EFEEDS_DB_ERR, "unable to hide feed", er)
		return e
	}

	return nil
}

func (fr *feedRecipes) accAdapter(ctx context.Context, t mqAcc.AccountTask, res **mqAcc.AccountRes) errro.Error {
	if err := fr.accContr.Publish(ctx, t); err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
		return e
	}

	err := fr.accContr.Consume(ctx, res)
	if err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
		return e
	}

	if (*res).Code != errro.SUCCESS {
		e := errro.New((*res).Code, "failed to lookup user")
		return e
	}

	return nil
}
