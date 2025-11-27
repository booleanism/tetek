package contracts

import (
	"context"
	"fmt"

	mqAcc "github.com/booleanism/tetek/account/amqp"
	mqAuth "github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/errro"
)

func AccAdapter(ctx context.Context, acc AccountDealer, task mqAcc.AccountTask, res **mqAcc.AccountResult) errro.Error {
	return adapter(ctx, acc, task, res)
}

func AuthAdapter(ctx context.Context, auth AuthDealer, task mqAuth.AuthTask, res **mqAuth.AuthResult) errro.Error {
	return adapter(ctx, auth, task, res)
}

func FeedsAdapter(ctx context.Context, feeds FeedsDealer, task mqAuth.AuthTask, res **mqAuth.AuthResult) errro.Error {
	return adapter(ctx, feeds, task, res)
}

func CommentsAdapter(ctx context.Context, comms FeedsDealer, task mqAuth.AuthTask, res **mqAuth.AuthResult) errro.Error {
	return adapter(ctx, comms, task, res)
}

func adapter(ctx context.Context, d any, t any, r any) errro.Error {
	if err := d.(Dealer).Publish(ctx, t); err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, fmt.Sprintf("failed to publish %s task", d.(Dealer).Name()), err)
		return e
	}

	err := d.(Dealer).Consume(ctx, r)
	if err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, fmt.Sprintf("failed consuming %s result", d.(Dealer).Name()), err)
		return e
	}
	return nil
}
