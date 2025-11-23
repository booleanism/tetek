package repo

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (r *feedsRepo) HideFeed(ctx context.Context, hf **model.HiddenFeed) error {
	ctx, log := loggr.GetLogger(ctx, "hideFeed-repo")
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer q.Release()

	f := model.HiddenFeed{}
	err = q.QueryRow(
		ctx,
		"INSERT INTO hiddenfeeds (id, to_uname, feed) VALUES ($1, $2, $3) ON CONFLICT (to_uname, feed) DO UPDATE SET feed = EXCLUDED.feed RETURNING id, to_uname, feed",
		&(*hf).ID, &(*hf).To, &(*hf).FeedID,
	).Scan(&f.ID, &f.To, &f.FeedID)
	if err != nil {
		log.Error(err, "failed to set feeds state into hiddenfeeds")
	}

	return err
}
