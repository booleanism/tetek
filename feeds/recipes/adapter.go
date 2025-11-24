package recipes

import (
	"context"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/pkg/errro"
)

func (fr feedRecipes) commAdapter(ctx context.Context, t amqp.CommentsTask, res **amqp.CommentsResult) errro.Error {
	if err := fr.commContr.Publish(ctx, t); err != nil {
		e := errro.FromError(errro.ErrCommPubFail, "failed to publish comments task", err)
		return e
	}

	if err := fr.commContr.Consume(ctx, res); err != nil {
		e := errro.FromError(errro.ErrCommConsumeFail, "failed to consume comments task", err)
		return e
	}

	return nil
}
