package scanner

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/loggr"
)

type ScanFn[T any] = func(ctx context.Context, rws db.Rows, buf *[]T, logKeysAndValues ...any) (n int, err error)

func CommentsScanner(ctx context.Context, rws db.Rows, buf *[]entities.Comment, logKeysAndValues ...any) (n int, err error) {
	_, log := loggr.GetLogger(ctx, "commentsScanner")
	n = 0
	for rws.Next() {
		f := entities.Comment{}
		if err = rws.Scan(&f.ID, &f.Parent, &f.Text, &f.By, &f.CreatedAt); err != nil {
			log.Error(err, "error occours scanning comments rows", "error", err)
			return n, err
		}
		*buf = append(*buf, f)
		n++
	}

	err = rws.Err()
	if err != nil {
		log.Error(err, "error occours after scanning comments rows", "error", err)
		return n, err
	}

	if n == 0 {
		log.V(1).Info("no errors, but zero rows comments returned", logKeysAndValues...)
		return 0, sql.ErrNoRows
	}

	return n, nil
}
