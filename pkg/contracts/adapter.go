package contracts

import (
	"context"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/pkg/errro"
)

func AccAdapter(ctx context.Context, acc AccountSubscribe, task amqp.AccountTask, res **amqp.AccountResult) errro.Error {
	if err := acc.Publish(ctx, task); err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, "failed to publish account task", err)
		return e
	}

	err := acc.Consume(ctx, res)
	if err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, "failed consuming account result", err)
		return e
	}
	return nil
}
