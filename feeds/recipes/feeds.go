package recipes

import (
	"context"
	"errors"
	"strings"

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

	f, er := fr.repo.Feeds(ctx, ff)
	if er != nil {
		var pgErr *pgconn.PgError
		if !errors.As(er, &pgErr) {
			e := errro.FromError(errro.ErrFeedsDBError, "error fetching feeds", er)
			return nil, e
		}

		if pgErr.Code == "23505" {
			e := errro.New(errro.ErrFeedsNoFeeds, "no feed(s) found")
			return nil, e
		}
	}

	if len(f) == 0 {
		e := errro.New(errro.ErrFeedsNoFeeds, "no feed(s) found")
		return nil, e
	}

	return f, nil
}

func (fr *feedRecipes) New(ctx context.Context, req NewFeedRequest) errro.Error {
	rFeed := req.ToFeed()
	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)

	rFeed.By = jwt.Claims.Uname
	rFeed.Type = strings.ToUpper(rFeed.Type)

	_, er := fr.repo.NewFeed(ctx, rFeed)
	if er != nil {
		e := errro.FromError(errro.ErrFeedsDBError, "could not insert new feed", er)
		return e
	}

	return nil
}

func (fr *feedRecipes) Delete(ctx context.Context, req DeleteRequest) errro.Error {
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

	fDel, er := fr.repo.DeleteFeed(ctx, ff)
	if er == nil {
		return nil
	}

	if er == pgx.ErrNoRows {
		e := errro.New(errro.ErrFeedsNoFeeds, "failed to delete feed: no row")
		return e
	}

	if fDel.DeletedAt == nil {
		e := errro.New(errro.ErrFeedsNoFeeds, "no such feed")
		return e
	}

	e := errro.New(errro.ErrFeedsDBError, "somthing happen when trying to delete feed")
	return e
}

// TODO:
// Validate feed by req.Id
func (fr *feedRecipes) Hide(ctx context.Context, req HideRequest) errro.Error {
	jwt := ctx.Value(keystore.AuthRes{}).(*mqAuth.AuthResult)
	hf := repo.HiddenFeeds{
		ID:     uuid.NewString(),
		To:     jwt.Claims.Uname,
		FeedID: req.ID,
	}

	_, er := fr.repo.HideFeed(ctx, hf)
	if er != nil {
		e := errro.FromError(errro.ErrFeedsDBError, "unable to hide feed", er)
		return e
	}

	return nil
}
