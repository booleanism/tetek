package repo

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedsFilter struct {
	Type     string `json:"type"`
	Order    int    `json:"order"` // 0: by points, 1: newest, 2: oldest
	Id       string `json:"id"`
	Offset   uint64 `json:"offset"`
	By       string `json:"by"`
	HiddenTo string `json:"hidden_to"`
}

const limit = 30

type FeedsRepo interface {
	Feeds(context.Context, FeedsFilter) ([]model.Feed, error)
	NewFeed(context.Context, model.Feed) (model.Feed, error)
	DeleteFeed(context.Context, model.Feed) (model.Feed, error)
	HideFeed(context.Context, HiddenFeeds) (HiddenFeeds, error)
}

type feedsRepo struct {
	pool *pgxpool.Pool
	sq   *squirrel.StatementBuilderType
}

func New(pool *pgxpool.Pool, sq *squirrel.StatementBuilderType) *feedsRepo {
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
	for rws.Next() {
		f := model.Feed{}
		if err := rws.Scan(&f.Id, &f.Title, &f.Url, &f.Text, &f.By, &f.Type, &f.Points, &f.NCommnents, &f.Created_At); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
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
		feed.Id, feed.Title, feed.Url, feed.Text, feed.By, feed.Type, feed.Points, feed.NCommnents, feed.Created_At,
	).Scan(&f.Id, &f.Title, &f.Url, &f.Text, &f.By, &f.Type, &f.Points, &f.NCommnents, &f.Created_At, &f.Deleted_At)

	return f, err
}

func (r *feedsRepo) DeleteFeed(ctx context.Context, feed model.Feed) (model.Feed, error) {
	q, err := r.pool.Acquire(ctx)
	if err != nil {
		return model.Feed{}, err
	}
	defer q.Release()

	base := r.sq.Update("feeds").Set("deleted_at", time.Now())

	where := squirrel.Eq{"id": feed.Id}
	if feed.By != "" {
		where["by"] = feed.By
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
	).Scan(&f.Deleted_At)

	return f, err
}
