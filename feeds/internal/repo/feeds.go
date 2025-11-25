package repo

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FeedsFilter struct {
	Type     string    `json:"type"`
	Order    int       `json:"order"` // 0: by points, 1: newest, 2: oldest
	ID       uuid.UUID `json:"id"`
	Offset   uint64    `json:"offset"`
	By       string    `json:"by"`
	HiddenTo string    `json:"hidden_to"`
}

const limit = 30

type FeedsRepo interface {
	Feeds(context.Context, FeedsFilter, *[]model.Feed) error
	NewFeed(context.Context, **model.Feed) error
	DeleteFeed(context.Context, FeedsFilter, **model.Feed) error
	HideFeed(context.Context, **model.HiddenFeed) error
}

type feedsRepo struct {
	pool db.Acquireable
	sq   squirrel.StatementBuilderType
}

func New(pool db.Acquireable, sq squirrel.StatementBuilderType) *feedsRepo {
	return &feedsRepo{pool, sq}
}

func (r *feedsRepo) Feeds(ctx context.Context, ff FeedsFilter, feedsBuf *[]model.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "getFeeds-repo")
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer q.Release()

	base := r.sq.Select("f.id", "f.title", "f.url", "f.text", "f.by", "f.type", "f.points", "f.n_comments", "f.created_at").From("feeds f")

	pred := map[string]any{
		"f.deleted_at": nil,
	}
	if ff.HiddenTo != "" {
		base = base.LeftJoin("hiddenfeeds hf on f.id = hf.feed AND hf.to_uname = ?", ff.HiddenTo)
		pred["hf.id"] = nil
	}

	if ff.ID.String() != "00000000-0000-0000-0000-000000000000" {
		pred["f.id"] = ff.ID
	}

	base = base.Where(pred)

	if ff.Type != "" {
		base = base.Where(squirrel.Eq{"f.type": ff.Type})
	}

	order := "f.points DESC"
	if ff.Order == 0 {
		order = "f.created_at ASC"
	}
	if ff.Order == 1 {
		order = "f.created_at DESC"
	}

	base = base.OrderBy(order).Limit(limit)

	if ff.Offset != 0 {
		base = base.Offset(ff.Offset * limit)
	}

	sql, args, err := base.ToSql()
	if err != nil {
		log.Error(err, "failed to build sql queries", "filter", ff)
		return err
	}

	rws, err := q.Query(ctx, sql, args...)
	if err != nil {
		log.Error(err, "failed to execute sql queries", "filter", ff)
		return err
	}

	defer rws.Close()
	return scanRows(ctx, rws, ff, feedsBuf)
}

func scanRows(ctx context.Context, rws pgx.Rows, ff FeedsFilter, buf *[]model.Feed) error {
	_, log := loggr.GetLogger(ctx, "feedsScanner-repo")
	// feeds := []model.Feed{}
	n := 0
	for rws.Next() {
		f := model.Feed{}
		if err := rws.Scan(&f.ID, &f.Title, &f.URL, &f.Text, &f.By, &f.Type, &f.Points, &f.NCommnents, &f.CreatedAt); err != nil {
			log.Error(err, "error occours scanning feeds rows")
			return err
		}
		*buf = append(*buf, f)
		n++
	}

	if n == 0 {
		log.V(1).Info("zero row feeds", "filter", ff)
		return pgx.ErrNoRows
	}

	err := rws.Err()
	if err != nil {
		log.Error(err, "error occours after scanning feeds rows")
		return err
	}

	return nil
}

func (r *feedsRepo) NewFeed(ctx context.Context, feed **model.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "newFeed-repo")
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer q.Release()

	err = q.QueryRow(ctx, "INSERT INTO feeds (id, title, url, text, by, type, points, n_comments, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *",
		&(*feed).ID, &(*feed).Title, &(*feed).URL, &(*feed).Text, &(*feed).By, &(*feed).Type, &(*feed).Points, &(*feed).NCommnents, &(*feed).CreatedAt,
	).Scan(&(*feed).ID, &(*feed).Title, &(*feed).URL, &(*feed).Text, &(*feed).By, &(*feed).Type, &(*feed).Points, &(*feed).NCommnents, &(*feed).CreatedAt, &(*feed).DeletedAt)
	if err != nil {
		log.Error(err, "failed to insert into feeds", "model", feed)
	}

	return err
}

func (r *feedsRepo) DeleteFeed(ctx context.Context, ff FeedsFilter, feed **model.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "deleteFeed-repo")
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer q.Release()

	base := r.sq.Update("feeds").Set("deleted_at", time.Now())

	where := squirrel.Eq{"id": ff.ID}
	if ff.By != "" {
		where["by"] = ff.By
	}

	base = base.Where(where)
	final := base.Suffix("RETURNING deleted_at")
	sql, args, err := final.ToSql()
	if err != nil {
		log.Error(err, "failed to build sql queries", "filter", ff)
		return err
	}

	err = q.QueryRow(
		ctx, sql, args...,
	).Scan(&(*feed).DeletedAt)
	if err != nil {
		log.Error(err, "failed to update feed state into deleted feed", "filter", ff)
	}

	return err
}

func (r *feedsRepo) HideFeed(ctx context.Context, hf **model.HiddenFeed) error {
	ctx, log := loggr.GetLogger(ctx, "hideFeed-repo")
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer q.Release()

	err = q.QueryRow(
		ctx,
		"INSERT INTO hiddenfeeds (id, to_uname, feed) VALUES ($1, $2, $3) ON CONFLICT (to_uname, feed) DO UPDATE SET feed = EXCLUDED.feed RETURNING id, to_uname, feed",
		&(*hf).ID, &(*hf).To, &(*hf).FeedID,
	).Scan(&(*hf).ID, &(*hf).To, &(*hf).FeedID)
	if err != nil {
		log.Error(err, "failed to set feeds state into hiddenfeeds")
	}

	return err
}
