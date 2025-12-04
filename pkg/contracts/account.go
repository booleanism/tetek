package contracts

import (
	"context"

	messaging "github.com/booleanism/tetek/account/infra/messaging/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type AccountDealer interface {
	Dealer
}

type localAccContr struct {
	d DealContract[messaging.AccountTask, messaging.AccountResult]
}

func AccountAssent(con *amqp091.Connection) *localAccContr {
	return &localAccContr{
		d: DealContract[messaging.AccountTask, messaging.AccountResult]{
			con:      con,
			name:     "account",
			exchange: messaging.AccountExchange,
			trk:      messaging.AccountTaskRk,
			rrk:      messaging.AccountResRk,
			res:      make(map[string]chan *messaging.AccountResult),
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
	return c.d.resListener(ctx, name, messaging.AccountSetup)
}

func (c *localAccContr) Name() string {
	return c.d.name
}
