package hiddenfeeds

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (fh feedsHidder) GetHidden(ctx context.Context, by string, buf *[]entities.HiddenFeed) (int, error) {
	ctx, log := loggr.GetLogger(ctx, "getHidden-repo")
	query := `SELECT id, to_uname, feed FROM hiddenfeeds WHERE to_uname = $1`
	rws, err := fh.exec(ctx, fh.db, query, by)
	if err != nil {
		log.Error(err, "failed to set feeds state into hiddenfeeds")
	}

	return fh.scan(ctx, rws, buf, "by", by)
}
