package repo

import (
	"context"

	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const Limit = 30

type CommentsRepo interface {
	GetComments(context.Context, CommentFilter, *[]model.Comment) (int, error)
	NewComment(context.Context, **model.Comment) error
}

type CommentFilter struct {
	ID     uuid.UUID
	Head   uuid.UUID
	By     string
	Offset int
}

type commRepo struct {
	db db.Acquireable
}

func NewCommRepo(db db.Acquireable) commRepo {
	return commRepo{db}
}

func (c commRepo) GetComments(ctx context.Context, ff CommentFilter, com *[]model.Comment) (int, error) {
	ctx, log := loggr.GetLogger(ctx, "getComments-repo")
	d, err := c.db.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer d.Release()

	q := `
	WITH RECURSIVE flat_comments AS (
		SELECT id, parent, text, by, created_at 
			FROM Comments
				WHERE parent = $1 OR id = $1
		UNION ALL
		SELECT c.id, c.parent, c.text, c.by, c.created_at 
			FROM Comments c 
			JOIN flat_comments ct 
				ON ct.id = c.parent
	)
	SELECT c.id, c.parent, c.text, c.by, c.created_at
		FROM flat_comments c 
	LIMIT $2 OFFSET $3;`

	rws, err := d.Query(ctx, q, ff.Head, Limit, ff.Offset)
	if err != nil {
		e := errro.FromError(errro.ErrCommQueryError, "failed execute query", err)
		log.Error(err, e.Error(), q, "$1", ff.Head, "$2", Limit, "$3", ff.Offset)
		return 0, e
	}

	n := 0
	for rws.Next() {
		c := model.Comment{}
		err := rws.Scan(&c.ID, &c.Parent, &c.Text, &c.By, &c.CreatedAt)
		if err != nil {
			e := errro.FromError(errro.ErrCommScanError, "error while scanning partial rows", err)
			log.Error(err, e.Error(), "row-index", n)
			return 0, e
		}
		n++
		(*com) = append((*com), c)
	}

	err = rws.Err()
	if err != nil {
		e := errro.FromError(errro.ErrCommScanError, "found error after scanning", err)
		log.Error(err, e.Error(), "rows", n)
		return 0, e
	}

	if n == 0 {
		log.V(1).Info("zero row comments", "filter", ff)
		return 0, pgx.ErrNoRows
	}

	return n, nil
}

func (c commRepo) NewComment(ctx context.Context, com **model.Comment) error {
	ctx, log := loggr.GetLogger(ctx, "newComment-repo")
	d, err := c.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer d.Release()

	sql := `INSERT INTO comments (
		id, parent, text, by, created_at
	) VALUES ($1, $2, $3, $4, $5) RETURNING *`

	r := d.QueryRow(ctx, sql, (*com).ID, (*com).Parent, (*com).Text, (*com).By, (*com).CreatedAt)
	err = r.Scan(&(*com).ID, &(*com).Parent, &(*com).Text, &(*com).By, &(*com).CreatedAt)
	if err != nil {
		e := errro.FromError(errro.ErrCommInsertError, "failed to insert comment", err)
		log.Error(err, "failed to insert comment", "query", sql, "data", com)
		return e
	}

	return nil
}
