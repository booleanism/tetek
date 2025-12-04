package hiddenfeeds

import (
	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type feedsHidder struct {
	db   db.DB
	exec db.QueryExecutorFn
	scan scanner.ScanFn[entities.HiddenFeed]
}

func NewFeedsHiderRepo(db db.DB) feedsHidder {
	return feedsHidder{db, nil, nil}
}

func (fh feedsHidder) SetScanner(scanFn scanner.ScanFn[entities.HiddenFeed]) feedsHidder {
	fh.scan = scanFn
	return fh
}

func (fh feedsHidder) SetExecutor(execFn db.QueryExecutorFn) feedsHidder {
	fh.exec = execFn
	return fh
}
