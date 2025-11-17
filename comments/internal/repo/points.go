package repo

import (
	"context"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
)

func (r commRepo) Upvote(ctx context.Context, cf CommentFilter) error {
	ctx, log := loggr.GetLogger(ctx, "repo/upvote")
	d, err := r.Acquire(ctx)
	if err != nil {
		e := errro.FromError(errro.ECOMM_DB_ERR, "failed to acquire db pool", err)
		log.Error(e, e.Error())
		return e
	}

	sql := "INSERT INTO points (id, by, comments_id) VALUES ($1, $2, $3) ON CONFLICT (by, comments_id) DO NOTHING;"
	_, err = d.Exec(ctx, sql, uuid.New(), cf.By, cf.Id)
	if err != nil {
		e := errro.FromError(errro.ECOMM_UPVOTE_FAIL, "failed to upvote comments", err)
		log.Error(err, e.Error(), "filter", cf)
		return e
	}

	return nil
}

func (r commRepo) Downvote(ctx context.Context, cf CommentFilter) error {
	ctx, log := loggr.GetLogger(ctx, "repo/downvote")
	d, err := r.Acquire(ctx)
	if err != nil {
		e := errro.FromError(errro.ECOMM_DB_ERR, "failed to acquire db pool", err)
		log.Error(e, e.Error())
		return e
	}

	sql := "DELETE FROM points WHERE by = $1 AND comments_id = $2;"
	_, err = d.Exec(ctx, sql, cf.By, cf.Id)
	if err != nil {
		e := errro.FromError(errro.ECOMM_DOWNVOTE_FAIL, "failed to downvote comments", err)
		log.Error(err, e.Error(), "filter", cf)
		return e
	}

	return nil
}
