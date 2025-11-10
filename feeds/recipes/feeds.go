package recipes

import (
	"context"
	"errors"
	"strings"

	mqAcc "github.com/booleanism/tetek/account/amqp"
	mqAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type FeedRecipes interface {
	Feeds(context.Context, repo.FeedsFilter, *mqAuth.AuthResult) ([]model.Feed, errro.Error)
	New(context.Context, model.Feed, *mqAuth.AuthResult) errro.Error
	Delete(context.Context, repo.FeedsFilter, *mqAuth.AuthResult) errro.Error
	Hide(context.Context, repo.FeedsFilter, *mqAuth.AuthResult) errro.Error
}

type feedRecipes struct {
	repo     repo.FeedsRepo
	accContr *contract.LocalAccContr
}

func NewRecipes(repo repo.FeedsRepo, accContr *contract.LocalAccContr) *feedRecipes {
	return &feedRecipes{repo, accContr}
}

func (fr *feedRecipes) Feeds(ctx context.Context, ff repo.FeedsFilter, jwt *mqAuth.AuthResult) ([]model.Feed, errro.Error) {
	traceId := ctx.Value(helper.RequestIdKey{}).(string)
	res := &mqAcc.AccountRes{}
	if jwt != nil {
		t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
		r, err := fr.accAdapter(traceId, t)
		if err != nil {
			return nil, err
		}
		res = r
	}

	if ff.Type == "" {
		ff.Type = "M"
	}
	ff.Type = strings.ToUpper(ff.Type)
	ff.HiddenTo = res.Detail.Uname

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

func (fr *feedRecipes) New(ctx context.Context, rFeed model.Feed, jwt *mqAuth.AuthResult) errro.Error {
	traceId := ctx.Value(helper.RequestIdKey{}).(string)
	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	res, err := fr.accAdapter(traceId, t)
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

func (fr *feedRecipes) Delete(ctx context.Context, ff repo.FeedsFilter, jwt *mqAuth.AuthResult) errro.Error {
	traceId := ctx.Value(helper.RequestIdKey{}).(string)
	if ff.Id.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.EFEEDS_MISSING_REQUIRED_FIELD, "missing required field")
		return e
	}

	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	res, err := fr.accAdapter(traceId, t)
	if err != nil {
		return err
	}

	f := repo.FeedsFilter{
		Id: ff.Id,
		By: res.Detail.Uname,
	}

	fDel, er := fr.repo.DeleteFeed(ctx, f)
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

func (fr *feedRecipes) Hide(ctx context.Context, ff repo.FeedsFilter, jwt *mqAuth.AuthResult) errro.Error {
	traceId := ctx.Value(helper.RequestIdKey{}).(string)
	if ff.Id.String() == "00000000-0000-0000-0000-000000000000" || ff.HiddenTo == "" {
		e := errro.New(errro.EFEEDS_MISSING_REQUIRED_FIELD, "missing required field")
		return e
	}

	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	res, err := fr.accAdapter(traceId, t)
	if err != nil {
		return err
	}

	hf := repo.HiddenFeeds{
		Id:     uuid.NewString(),
		To:     res.Detail.Uname,
		FeedId: ff.Id,
	}

	_, er := fr.repo.HideFeed(ctx, hf)
	if er != nil {
		e := errro.FromError(errro.EFEEDS_DB_ERR, "unable to hide feed", er)
		return e
	}

	return nil
}

func (fr *feedRecipes) accAdapter(corrId string, t mqAcc.AccountTask) (*mqAcc.AccountRes, errro.Error) {
	if err := fr.accContr.Publish(corrId, t); err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
		return nil, e
	}

	res, err := fr.accContr.Consume(corrId)
	if err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
		return nil, e
	}

	if res.Code != errro.SUCCESS {
		e := errro.New(res.Code, "failed to lookup user")
		return nil, e
	}

	return res, nil
}
