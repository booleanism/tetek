package contracts

import (
	"context"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/rabbitmq/amqp091-go"
)

type AccountDealer interface {
	Dealer
}

type localAccContr struct {
	d DealContract[amqp.AccountTask, amqp.AccountResult]
}

func AccountAssent(con *amqp091.Connection) *localAccContr {
	return &localAccContr{
		d: DealContract[amqp.AccountTask, amqp.AccountResult]{
			con:      con,
			name:     "account",
			exchange: amqp.AccountExchange,
			trk:      amqp.AccountTaskRk,
			rrk:      amqp.AccountResRk,
			res:      make(map[string]chan *amqp.AccountResult),
		},
	}
}

func (c *localAccContr) Publish(ctx context.Context, task any) error {
	return c.d.publish(ctx, task)
}

func (c *localAccContr) Consume(ctx context.Context, res any) error {
	return c.d.consume(ctx, res)
}

func (c *localAccContr) AccountResListener(ctx context.Context, name string) error {
	return c.d.resListener(ctx, name, amqp.AccountSetup)
}

func (c *localAccContr) Name() string {
	return c.d.name
}
