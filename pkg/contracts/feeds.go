package contracts

import (
	"context"

	"github.com/booleanism/tetek/feeds/infra/amqp"
	"github.com/rabbitmq/amqp091-go"
)

type FeedsDealer interface {
	Dealer
}

type localFeedsContr struct {
	d DealContract[amqp.FeedsTask, amqp.FeedsResult]
}

func FeedsAssent(con *amqp091.Connection) *localFeedsContr {
	return &localFeedsContr{
		d: DealContract[amqp.FeedsTask, amqp.FeedsResult]{
			con:      con,
			name:     "feeds",
			exchange: amqp.FeedsExchange,
			trk:      amqp.FeedsTaskRk,
			rrk:      amqp.FeedsResRk,
			res:      make(map[string]chan *amqp.FeedsResult),
		},
	}
}

func (c *localFeedsContr) Publish(ctx context.Context, task any) error {
	return c.d.publish(ctx, task)
}

func (c *localFeedsContr) Consume(ctx context.Context, res any) error {
	return c.d.consume(ctx, res)
}

func (c *localFeedsContr) FeedsResListener(ctx context.Context, name string) error {
	return c.d.resListener(ctx, name, amqp.FeedsSetup)
}

func (c *localFeedsContr) Name() string {
	return c.d.name
}
