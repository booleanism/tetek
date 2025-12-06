package repo

import (
	"context"

	comments "github.com/booleanism/tetek/comments/internal/infra/db/sql"
	"github.com/booleanism/tetek/comments/internal/internal/domain/repo"
	"github.com/booleanism/tetek/comments/internal/usecases/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type CommentsRepo interface {
	repo.CommentsEntityRepo
}

type repoComments struct {
	repo.CommentsEntityRepo
}

func defaultExecutor(ctx context.Context, db db.DB, query string, args ...any) (db.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func NewCommentsRepo(d db.DB) CommentsRepo {
	return repoComments{
		setupComments(d),
	}
}

func setupComments(d db.DB) repo.CommentsEntityRepo {
	f := comments.NewCommentsRepo(d)
	f = f.SetExecutor(defaultExecutor)
	f = f.SetScanner(scanner.CommentsScanner)
	return f
}
