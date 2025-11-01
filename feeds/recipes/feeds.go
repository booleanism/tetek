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
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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
	res := &mqAcc.AccountRes{}
	if jwt != nil {
		id := uuid.NewString()
		r, err := fr.accAdapter(id, jwt)
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
			return nil, loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
				e := errro.New(errro.EFEEDS_DB_ERR, "error fetching feeds")
				z.Error(er, e.Error(), "filter", ff)
				return e
			})
		}

		if pgErr.Code == "23505" {
			return nil, loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
				e := errro.New(errro.EFEEDS_NO_FEEDS, "no feed(s) found")
				z.Error(pgErr, e.Error(), "filter", ff)
				return e
			})
		}
	}

	if len(f) == 0 {
		return nil, loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_NO_FEEDS, "no feed(s) found")
			z.Error(e, e.Error(), "filter", ff)
			return e
		})
	}

	return f, nil
}

func (fr *feedRecipes) New(ctx context.Context, rFeed model.Feed, jwt *mqAuth.AuthResult) errro.Error {
	id := uuid.NewString()
	res, err := fr.accAdapter(id, jwt)
	if err != nil {
		return err
	}

	rFeed.By = res.Detail.Uname
	rFeed.Type = strings.ToUpper(rFeed.Type)

	_, er := fr.repo.NewFeed(ctx, rFeed)
	if er != nil {
		return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_DB_ERR, "could not insert new feed")
			z.Error(er, e.Error(), "feed", rFeed)
			return e
		})
	}

	return nil
}

func (fr *feedRecipes) Delete(ctx context.Context, ff repo.FeedsFilter, jwt *mqAuth.AuthResult) errro.Error {
	if ff.Id == "" {
		return loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_MISSING_REQUIRED_FIELD, "missing required field")
			z.Error(e, "missing id", "filter", ff)
			return e
		})
	}

	id := uuid.NewString()
	res, err := fr.accAdapter(id, jwt)
	if err != nil {
		return err
	}

	f := model.Feed{
		Id: ff.Id,
		By: res.Detail.Uname,
	}

	fDel, er := fr.repo.DeleteFeed(ctx, f)
	if er == nil {
		return nil
	}

	if er == pgx.ErrNoRows {
		return loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_NO_FEEDS, "failed to delete feed: no row")
			z.Error(er, e.Error(), "feed", f)
			return e
		})
	}

	if fDel.Deleted_At == nil {
		return loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_NO_FEEDS, "failed to delete feed: nil")
			z.Error(er, e.Error(), "feed", f)
			return e
		})
	}

	return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
		e := errro.New(errro.EFEEDS_DB_ERR, "somthing happen when trying to delete feed")
		z.Error(er, e.Error(), "feed", f)
		return e
	})
}

func (fr *feedRecipes) Hide(ctx context.Context, ff repo.FeedsFilter, jwt *mqAuth.AuthResult) errro.Error {
	if ff.Id == "" || ff.HiddenTo == "" {
		return errro.New(errro.EFEEDS_MISSING_REQUIRED_FIELD, "either id or hiddento should not empty")
	}

	id := uuid.NewString()
	res, err := fr.accAdapter(id, jwt)
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
		return loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EFEEDS_DB_ERR, "unable to hide feed")
			z.Error(er, e.Error())
			return e
		})
	}

	return nil
}

func (fr *feedRecipes) accAdapter(corrId string, jwt *mqAuth.AuthResult) (*mqAcc.AccountRes, errro.Error) {
	t := mqAcc.AccountTask{Cmd: 0, User: mqAcc.User{Uname: jwt.Claims.Uname}}
	if err := fr.accContr.Publish(corrId, t); err != nil {
		return nil, loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
			z.Error(err, e.Error(), "id", corrId, "user", jwt.Claims.Uname)
			return e
		})
	}

	res, err := fr.accContr.Consume(corrId)
	if err != nil {
		return nil, loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to communicate with account service")
			z.Error(err, e.Error(), "id", corrId, "user", jwt.Claims.Uname)
			return e
		})
	}

	if res.Code != errro.SUCCESS {
		return nil, loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
			e := errro.New(res.Code, "failed to lookup user")
			z.Error(e, e.Error(), "id", corrId, "user", jwt.Claims.Uname)
			return e
		})
	}

	return res, nil
}
