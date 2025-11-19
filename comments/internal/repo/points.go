package repo

import (
	"context"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
)

func (c commRepo) Upvote(ctx context.Context, cf CommentFilter) error {
	ctx, log := loggr.GetLogger(ctx, "repo/upvote")
	d, err := c.Acquire(ctx)
	if err != nil {
		e := errro.FromError(errro.ErrCommDBError, "failed to acquire db pool", err)
		log.Error(e, e.Error())
		return e
	}

	sql := "INSERT INTO points (id, by, comments_id) VALUES ($1, $2, $3) ON CONFLICT (by, comments_id) DO NOTHING;"
	_, err = d.Exec(ctx, sql, uuid.New(), cf.By, cf.ID)
	if err != nil {
		e := errro.FromError(errro.ErrCommUpvoteFail, "failed to upvote comments", err)
		log.Error(err, e.Error(), "filter", cf)
		return e
	}

	return nil
}

func (c commRepo) Downvote(ctx context.Context, cf CommentFilter) error {
	ctx, log := loggr.GetLogger(ctx, "repo/downvote")
	d, err := c.Acquire(ctx)
	if err != nil {
		e := errro.FromError(errro.ErrCommDBError, "failed to acquire db pool", err)
		log.Error(e, e.Error())
		return e
	}

	sql := "DELETE FROM points WHERE by = $1 AND comments_id = $2;"
	_, err = d.Exec(ctx, sql, cf.By, cf.ID)
	if err != nil {
		e := errro.FromError(errro.ErrCommDownvoteFail, "failed to downvote comments", err)
		log.Error(err, e.Error(), "filter", cf)
		return e
	}

	return nil
}
