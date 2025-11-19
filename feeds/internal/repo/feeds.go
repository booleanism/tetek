package repo

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/internal/model"
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
	Feeds(context.Context, FeedsFilter) ([]model.Feed, error)
	NewFeed(context.Context, model.Feed) (model.Feed, error)
	DeleteFeed(context.Context, FeedsFilter) (model.Feed, error)
	HideFeed(context.Context, HiddenFeeds) (HiddenFeeds, error)
}

type feedsRepo struct {
	pool db.Acquireable
	sq   squirrel.StatementBuilderType
}

func New(pool db.Acquireable, sq squirrel.StatementBuilderType) *feedsRepo {
	return &feedsRepo{pool, sq}
}

func (r *feedsRepo) Feeds(ctx context.Context, ff FeedsFilter) ([]model.Feed, error) {
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
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
	if ff.Order == 1 {
		order = "f.created_at ASC"
	}
	if ff.Order == 2 {
		order = "f.created_at DESC"
	}

	base = base.OrderBy(order).Limit(limit)

	if ff.Offset != 0 {
		base = base.Offset(ff.Offset * limit)
	}

	sql, args, err := base.ToSql()
	if err != nil {
		return nil, err
	}

	rws, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	defer rws.Close()
	return scanRows(rws)
}

func scanRows(rws pgx.Rows) ([]model.Feed, error) {
	feeds := []model.Feed{}
	n := 0
	for rws.Next() {
		f := model.Feed{}
		if err := rws.Scan(&f.ID, &f.Title, &f.URL, &f.Text, &f.By, &f.Type, &f.Points, &f.NCommnents, &f.CreatedAt); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
		n++
	}

	if n == 0 {
		return nil, pgx.ErrNoRows
	}

	return feeds, nil
}

func (r *feedsRepo) NewFeed(ctx context.Context, feed model.Feed) (model.Feed, error) {
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return model.Feed{}, err
	}
	defer q.Release()

	f := model.Feed{}
	err = q.QueryRow(ctx, "INSERT INTO feeds (id, title, url, text, by, type, points, n_comments, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *",
		feed.ID, feed.Title, feed.URL, feed.Text, feed.By, feed.Type, feed.Points, feed.NCommnents, feed.CreatedAt,
	).Scan(&f.ID, &f.Title, &f.URL, &f.Text, &f.By, &f.Type, &f.Points, &f.NCommnents, &f.CreatedAt, &f.DeletedAt)

	return f, err
}

func (r *feedsRepo) DeleteFeed(ctx context.Context, ff FeedsFilter) (model.Feed, error) {
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return model.Feed{}, err
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
		return model.Feed{}, err
	}

	f := model.Feed{}
	err = q.QueryRow(
		ctx, sql, args...,
	).Scan(&f.DeletedAt)

	return f, err
}
