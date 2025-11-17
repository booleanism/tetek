package recipes

import (
	"context"

	amqpFeeds "github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (cr commRecipes) feedsAdapter(ctx context.Context, t amqpFeeds.FeedsTask, res **amqpFeeds.FeedsResult) errro.Error {
	_, log := loggr.GetLogger(ctx, "feeds-adapter")
	if err := cr.feeds.Publish(ctx, t); err != nil {
		e := errro.FromError(errro.ECOMM_PUB_FAIL, "failed to publish feeds task", err)
		log.Error(err, e.Error(), "task", t)
		return e
	}

	err := cr.feeds.Consume(ctx, res)
	if err != nil {
		e := errro.FromError(errro.ECOMM_CONSUME_FAIL, "failed to consume feeds result", err)
		log.Error(err, err.Error(), "task", t)
		return e
	}
	return nil
}
