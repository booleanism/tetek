package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/jackc/pgx/v5"
)

type ProfileRecipes interface {
	Profile(context.Context, ProfileRequest) (model.User, errro.Error)
}

type profileRecipe struct {
	repo repo.UserRepo
}

func (r profileRecipe) Profile(ctx context.Context, req ProfileRequest) (model.User, errro.Error) {
	u := &model.User{Uname: req.Uname}
	err := r.repo.GetUser(ctx, &u)
	if err == nil {
		return model.User{
			Uname:     u.Uname,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		}, nil
	}

	if err == pgx.ErrNoRows {
		e := errro.New(errro.ErrAccountNoUser, "couldn't find user")
		return model.User{}, e
	}

	e := errro.New(errro.ErrAccountDBError, "something happen with our database")
	return model.User{}, e
}
