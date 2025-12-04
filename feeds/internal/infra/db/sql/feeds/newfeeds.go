package feeds

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (fr feedsRepo) PutIntoFeed(ctx context.Context, feed **entities.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "putIntoFeed-repo")
	sqlStmt := `
	INSERT INTO feeds 
		(id, title, url, text, by, type, points, n_comments, created_at)
	VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING
		id, title, url, text, by, created_at;`

	rws, err := fr.exec(
		ctx,
		fr.db,
		sqlStmt,
		&(*feed).ID, &(*feed).Title, &(*feed).URL, &(*feed).Text, &(*feed).By, &(*feed).Type, &(*feed).Points, &(*feed).NCommnents, &(*feed).CreatedAt,
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

	buf := []entities.Feed{**feed}
	if _, err := fr.scan(ctx, rws, &buf, "feed", feed); err != nil {
		return err
	}
	(*feed) = &buf[0]

	return nil
}
