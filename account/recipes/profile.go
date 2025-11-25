package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/jackc/pgx/v5"
)

type ProfileRecipes interface {
	Profile(context.Context, ProfileRequest, **model.User) errro.Error
}

type profileRecipe struct {
	repo repo.UserRepo
}

func (r profileRecipe) Profile(ctx context.Context, req ProfileRequest, u **model.User) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "login-recipes")
	if req.Uname == "" {
		e := errro.New(errro.ErrAccountEmptyParam, "uname should not empty")
		log.V(1).Info(e.Msg())
		return e
	}

	err := r.repo.GetUser(ctx, u)
	if err == nil {
		return nil
	}

	if err == pgx.ErrNoRows {
		e := errro.New(errro.ErrAccountNoUser, "couldn't find user")
		return e
	}

	e := errro.New(errro.ErrAccountDBError, "something happen with our database")
	return e
}
