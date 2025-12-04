package feeds

import (
	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type feedsRepo struct {
	db   db.DB
	exec db.QueryExecutorFn
	scan scanner.ScanFn[entities.Feed]
}

func NewFeedsRepo(db db.DB) feedsRepo {
	return feedsRepo{db, nil, nil}
}

func (fr feedsRepo) SetScanner(scanFn scanner.ScanFn[entities.Feed]) feedsRepo {
	fr.scan = scanFn
	return fr
}

func (fr feedsRepo) SetExecutor(execFn db.QueryExecutorFn) feedsRepo {
	fr.exec = execFn
	return fr
}
