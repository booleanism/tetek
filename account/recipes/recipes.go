package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
)

type AccRecipes interface {
	RegistRecipes
	ProfileRecipes
}

type recipes struct {
	repo repo.UserRepo
}

func New(repo repo.UserRepo) *recipes {
	return &recipes{repo}
}

func (r *recipes) Profile(ctx context.Context, user model.User) (model.User, errro.Error) {
	return (&profileRecipe{r.repo}).Profile(ctx, user)
}

func (r *recipes) Regist(ctx context.Context, user model.User) errro.Error {
	return (&registRecipes{r.repo}).Regist(ctx, user)
}
