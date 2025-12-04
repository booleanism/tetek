package contracts

import (
	"context"

	messaging "github.com/booleanism/tetek/feeds/infra/messaging/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type FeedsDealer interface {
	Dealer
}

type localFeedsContr struct {
	d DealContract[messaging.FeedsTask, messaging.FeedsResult]
}

func FeedsAssent(con *amqp091.Connection) *localFeedsContr {
	return &localFeedsContr{
		d: DealContract[messaging.FeedsTask, messaging.FeedsResult]{
			con:      con,
			name:     "feeds",
			exchange: messaging.FeedsExchange,
			trk:      messaging.FeedsTaskRk,
			rrk:      messaging.FeedsResRk,
			res:      make(map[string]chan *messaging.FeedsResult),
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
	return c.d.resListener(ctx, name, messaging.FeedsSetup)
}

func (c *localFeedsContr) Name() string {
	return c.d.name
}
