package scanner

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/account/internal/domain/entities"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/loggr"
)

type ScanFn[T any] = func(ctx context.Context, rws db.Rows, buf *[]T, param ...any) (n int, err error)

func UserScanner(ctx context.Context, rws db.Rows, buf *[]entities.User, param ...any) (n int, err error) {
	_, log := loggr.GetLogger(ctx, "userScanner")
	n = 0
	for rws.Next() {
		u := entities.User{}
		if err = rws.Scan(&u.Uname, &u.Email, &u.Passwd, &u.Role); err != nil {
			log.Error(err, "error occours scanning user rows", "error", err)
			return n, err
		}
		*buf = append(*buf, u)
		n++
	}

	err = rws.Err()
	if err != nil {
		log.Error(err, "error occours after scanning user rows", "error", err)
		return n, err
	}

	if n == 0 {
		log.V(1).Info("no errors, but zero rows user returned")
		return 0, sql.ErrNoRows
	}

	return n, nil
}
