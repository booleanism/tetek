package db

import (
	"context"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	"github.com/booleanism/tetek/comments/internal/usecases/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/loggr"
)

type commRepo struct {
	db   db.DB
	scan scanner.ScanFn[entities.Comment]
	exec db.QueryExecutorFn
}

func NewCommentsRepo(db db.DB) commRepo {
	return commRepo{db, nil, nil}
}

func (cr commRepo) SetScanner(scanFn scanner.ScanFn[entities.Comment]) commRepo {
	cr.scan = scanFn
	return cr
}

func (cr commRepo) SetExecutor(execFn db.QueryExecutorFn) commRepo {
	cr.exec = execFn
	return cr
}

func (cr commRepo) PutComment(ctx context.Context, com **entities.Comment) error {
	ctx, log := loggr.GetLogger(ctx, "newComment-repo")

	query := `
	INSERT INTO comments 
		(id, parent, text, by, created_at) 
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING 
		id, parent, text, by, created_at;`

	rws, err := cr.exec(ctx, cr.db, query, (*com).ID, (*com).Parent, (*com).Text, (*com).By, (*com).CreatedAt)
	if err != nil {
		log.Error(err, "sql query execution failed", "query", query)
		return err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	return nil
}
