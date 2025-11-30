package adapter

import (
	"context"
	"fmt"

	mqAcc "github.com/booleanism/tetek/account/amqp"
	mqAuth "github.com/booleanism/tetek/auth/amqp"
	mqComm "github.com/booleanism/tetek/comments/amqp"
	mqFeeds "github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
)

func AccAdapter(ctx context.Context, acc contracts.AccountDealer, task mqAcc.AccountTask, res **mqAcc.AccountResult) errro.Error {
	return adapter(ctx, acc, task, res)
}

func AuthAdapter(ctx context.Context, auth contracts.AuthDealer, task mqAuth.AuthTask, res **mqAuth.AuthResult) errro.Error {
	return adapter(ctx, auth, task, res)
}

func FeedsAdapter(ctx context.Context, feeds contracts.FeedsDealer, task mqFeeds.FeedsTask, res **mqFeeds.FeedsResult) errro.Error {
	return adapter(ctx, feeds, task, res)
}

func CommentsAdapter(ctx context.Context, comms contracts.CommentsDealer, task mqComm.CommentsTask, res **mqComm.CommentsResult) errro.Error {
	return adapter(ctx, comms, task, res)
}

func adapter(ctx context.Context, d any, t any, r any) errro.Error {
	if err := d.(contracts.Dealer).Publish(ctx, t); err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, fmt.Sprintf("failed to publish %s task", d.(contracts.Dealer).Name()), err)
		return e
	}

	err := d.(contracts.Dealer).Consume(ctx, r)
	if err != nil {
		e := errro.FromError(errro.ErrAccountServiceUnavailable, fmt.Sprintf("failed consuming %s result", d.(contracts.Dealer).Name()), err)
		return e
	}
	return nil
}
