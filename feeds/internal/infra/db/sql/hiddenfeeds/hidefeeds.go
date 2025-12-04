package hiddenfeeds

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (fh feedsHidder) HideFeed(ctx context.Context, hf **entities.HiddenFeed) error {
	ctx, log := loggr.GetLogger(ctx, "hideFeed-repo")

	query := `
	INSERT INTO hiddenfeeds 
		(id, to_uname, feed) 
	VALUES ($1, $2, $3) 
	ON CONFLICT 
		(to_uname, feed) 
	DO 
		UPDATE SET feed = EXCLUDED.feed 
	RETURNING id, to_uname, feed;`

	rws, err := fh.exec(ctx, fh.db, query, &(*hf).ID, &(*hf).To, &(*hf).FeedID)
	if err != nil {
		log.Error(err, "failed to set feeds state into hiddenfeeds")
		return err
	}

	defer func() {
		if err := rws.Close(); err != nil {
			log.Error(err, "failed close rows")
		}
	}()

	buf := []entities.HiddenFeed{}
	if _, err := fh.scan(ctx, rws, &buf, "hiddenfeeds", hf); err != nil {
		return err
	}
	(*hf) = &buf[0]

	return nil
}
