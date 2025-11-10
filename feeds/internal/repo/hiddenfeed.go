package repo

import (
	"context"

	"github.com/google/uuid"
)

type HiddenFeeds struct {
	Id     string    `json:"id"`
	To     string    `json:"to"`
	FeedId uuid.UUID `json:"feed_id"`
}

func (r *feedsRepo) HideFeed(ctx context.Context, hf HiddenFeeds) (HiddenFeeds, error) {
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return HiddenFeeds{}, err
	}
	defer q.Release()

	f := HiddenFeeds{}
	err = q.QueryRow(
		ctx,
		"INSERT INTO hiddenfeeds (id, to_uname, feed) VALUES ($1, $2, $3) ON CONFLICT (to_uname, feed) DO UPDATE SET feed = EXCLUDED.feed RETURNING id, to_uname, feed",
		hf.Id, hf.To, hf.FeedId,
	).Scan(&f.Id, &f.To, &f.FeedId)

	return f, err
}
