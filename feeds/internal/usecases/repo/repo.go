package repo

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/repo"
	"github.com/booleanism/tetek/feeds/internal/domain/repo/scanner"
	"github.com/booleanism/tetek/feeds/internal/infra/db/sql/feeds"
	"github.com/booleanism/tetek/feeds/internal/infra/db/sql/hiddenfeeds"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type FeedsRepo interface {
	repo.FeedsEntityRepo
	repo.HiddenFeedsEntityRepo
}

type repoFeeds struct {
	repo.FeedsEntityRepo
	repo.HiddenFeedsEntityRepo
}

func defaultExecutor(ctx context.Context, db db.DB, query string, args ...any) (db.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func NewFeedsRepo(d db.DB) FeedsRepo {
	return repoFeeds{
		setupFeeds(d),
		setupHiddenFeeds(d),
	}
}

func setupFeeds(d db.DB) repo.FeedsEntityRepo {
	f := feeds.NewFeedsRepo(d)
	f = f.SetExecutor(defaultExecutor)
	f = f.SetScanner(scanner.FeedsScanner)
	return f
}

func setupHiddenFeeds(d db.DB) repo.HiddenFeedsEntityRepo {
	f := hiddenfeeds.NewFeedsHiderRepo(d)
	f = f.SetExecutor(defaultExecutor)
	f = f.SetScanner(scanner.HiddenFeedsScanner)
	return f
}
