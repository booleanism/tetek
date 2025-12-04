package scanner

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/jackc/pgx/v5"
)

// ScanFn is an function type signature that return n scanned rows.
// If error occours, you may need to check n.
// Don't forget to log properly.
type ScanFn[T any] = func(ctx context.Context, rws db.Rows, buf *[]T, logKeysAndValues ...any) (n int, err error)

func FeedsScanner(ctx context.Context, rws db.Rows, buf *[]entities.Feed, logKeysAndValues ...any) (n int, err error) {
	_, log := loggr.GetLogger(ctx, "feedsScanner")
	n = 0
	for rws.Next() {
		f := entities.Feed{}
		if err = rws.Scan(&f.ID, &f.Title, &f.Text, &f.URL, &f.By, &f.CreatedAt); err != nil {
			log.Error(err, "error occours scanning feeds rows", "error", err)
			return n, err
		}
		*buf = append(*buf, f)
		n++
	}

	err = rws.Err()
	if err != nil {
		log.Error(err, "error occours after scanning feeds rows", "error", err)
		return n, err
	}

	if n == 0 {
		log.V(1).Info("no errors, but zero rows feeds returned", logKeysAndValues...)
		return 0, sql.ErrNoRows
	}

	return n, nil
}

func HiddenFeedsScanner(ctx context.Context, rws db.Rows, buf *[]entities.HiddenFeed, logKeysAndValues ...any) (n int, err error) {
	_, log := loggr.GetLogger(ctx, "hiddenFeedsScanner")
	n = 0
	for rws.Next() {
		f := entities.HiddenFeed{}
		if err = rws.Scan(&f.ID, &f.To, &f.FeedID); err != nil {
			log.Error(err, "error occours scanning hiddenfeeds rows")
		}
		*buf = append(*buf, f)
		n++
	}

	err = rws.Err()
	if err != nil {
		log.Error(err, "error occours after scanning hiddenfeeds rows")
	}

	if n == 0 {
		log.V(1).Info("zero row feeds", logKeysAndValues...)
		err = pgx.ErrNoRows
	}

	return
}
