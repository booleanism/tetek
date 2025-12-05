package repo

import (
	"context"

	"github.com/booleanism/tetek/account/internal/domain/repo"
	"github.com/booleanism/tetek/account/internal/infra/db/sql/user"
	"github.com/booleanism/tetek/account/internal/usecases/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type UserRepo interface {
	repo.UserGetter
	repo.UserAdder
}

type repoUser struct {
	UserRepo
}

func defaultExecutor(ctx context.Context, db db.DB, query string, args ...any) (db.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func NewUserRepo(d db.DB) UserRepo {
	return repoUser{setupUserRepo(d)}
}

func setupUserRepo(d db.DB) UserRepo {
	f := user.NewUserRepo(d)
	f = f.SetExecutor(defaultExecutor)
	f = f.SetScanner(scanner.UserScanner)
	return f
}
