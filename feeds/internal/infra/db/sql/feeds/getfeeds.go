package feeds

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/google/uuid"
)

func (fr feedsRepo) GetFeedsByID(ctx context.Context, id uuid.UUID, buf **entities.Feed) error {
	ctx, log := loggr.GetLogger(ctx, "getFeedsByID-repo")
	sqlStmt := `
	SELECT 
		f.id, f.title, f.text, f.url, f.by, f.created_at 
		FROM feeds f
	WHERE 
		f.deleted_at IS NULL
	AND
		f.id = $1;`

	rws, err := fr.exec(ctx, fr.db, sqlStmt, id)
	if err != nil {
		log.Error(err, "failed to execute sql queries", "id", id)
		return err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	f := []entities.Feed{}
	_, err = fr.scan(ctx, rws, &f, "id", id)
	if err != nil {
		return err
	}
	(*buf) = &f[0]

	return nil
}

func (fr feedsRepo) GetFeedsByType(ctx context.Context, fType model.GetFeedsByType, buf *[]entities.Feed) (n int, err error) {
	ctx, log := loggr.GetLogger(ctx, "getFeedsByType-repo")
	sqlStmt := `
	SELECT 
		f.id, f.title, f.text, f.url, f.by, f.created_at 
	FROM feeds f
	WHERE 
		f.deleted_at IS NULL
	AND
		f.type = $1
	LIMIT $2
	OFFSET $3`

	rws, err := fr.exec(ctx, fr.db, sqlStmt, fType.Type, fType.Limit, fType.Offset)
	if err != nil {
		log.Error(err, "failed to execute sql queries", "type", fType)
		return 0, err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	return fr.scan(ctx, rws, buf, "type", fType)
}

func (fr feedsRepo) GetFeedsNotHiddenBy(ctx context.Context, fType model.GetFeedsByType, by string, buf *[]entities.Feed) (n int, err error) {
	ctx, log := loggr.GetLogger(ctx, "getFeedsNotHiddenBy-repo")
	sqlStmt := `
	SELECT 
		f.id, f.title, f.text, f.url, f.by, f.created_at 
	FROM feeds f
	LEFT JOIN hiddenfeeds hf
	ON
		f.id = hf.feed
	WHERE 
		f.deleted_at IS NULL
	AND
		f.type = $1
	AND 
		hf.id IS NULL
	LIMIT $2
	OFFSET $3`

	rws, err := fr.exec(ctx, fr.db, sqlStmt, fType.Type, fType.Limit, fType.Offset)
	if err != nil {
		log.Error(err, "failed to execute sql queries", "type", fType, "by", by)
		return 0, err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	return fr.scan(ctx, rws, buf, "type", fType)
}
