package db

import (
	"context"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
)

func (cr commRepo) GetComments(ctx context.Context, id uuid.UUID, lim, offset int, com *[]entities.Comment) (int, error) {
	ctx, log := loggr.GetLogger(ctx, "getComments-repo")

	query := `
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

	rws, err := cr.exec(ctx, cr.db, query, id, lim, offset)
	if err != nil {
		log.Error(err, "failed sql query executen", "query", query)
		return 0, err
	}
	defer rws.Close()

	return cr.scan(ctx, rws, com)
}
