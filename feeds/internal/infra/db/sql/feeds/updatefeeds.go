package feeds

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (fr feedsRepo) DeleteFeed(ctx context.Context, fd model.FeedDeletion, feed **entities.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "deleteFeed-repo")
	sqlStmt := `
	UPDATE feeds
	SET
		deleted_at = $1
	WHERE
		id = $2
	AND
		deleted_at IS NULL
	RETURNING
		id, title, text, url, by, created_at;`

	rws, err := fr.exec(
		ctx,
		fr.db,
		sqlStmt,
		fd.DeletedAt, fd.ID,
	)
	if err != nil {
		log.Error(err, "failed to execute sql queries", "feed", feed)
		return err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	buf := []entities.Feed{}
	if _, err := fr.scan(ctx, rws, &buf, "feed", feed); err != nil {
		return err
	}
	(*feed) = &buf[0]

	return nil
}
