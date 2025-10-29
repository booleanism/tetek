package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
)

type ProfileRecipes interface {
	Profile(context.Context, model.User) (model.User, errro.Error)
}

type profileRecipe struct {
	repo repo.UserRepo
}

func (r *profileRecipe) Profile(ctx context.Context, user model.User) (model.User, errro.Error) {
	u, err := r.repo.GetUser(ctx, user)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.User{}, loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				e := errro.New(errro.EACCOUNT_NO_USER, "couldn't find user")
				z.Error(err, e.Error(), "user", user)
				return e
			})
		}
		return model.User{}, loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_DB_ERR, "something happen with database interaction")
			z.Error(err, e.Error(), "user", user)
			return e
		})
	}

	return model.User{
		Uname:     u.Uname,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}
