package contracts

import (
	"context"

	messaging "github.com/booleanism/tetek/comments/infra/messaging/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type CommentsDealer interface {
	Dealer
}

type localCommContr struct {
	d DealContract[messaging.CommentsTask, messaging.CommentsResult]
}

func CommentsAssent(con *amqp091.Connection) *localCommContr {
	return &localCommContr{
		d: DealContract[messaging.CommentsTask, messaging.CommentsResult]{
			con:      con,
			name:     "comments",
			exchange: messaging.CommentsExchange,
			trk:      messaging.CommentsTaskRk,
			rrk:      messaging.CommentsResRk,
			res:      make(map[string]chan *messaging.CommentsResult),
		},
	}
}

func (c *localCommContr) Publish(ctx context.Context, task any) error {
	return c.d.publish(ctx, task)
}

func (c *localCommContr) Consume(ctx context.Context, res any) error {
	return c.d.consume(ctx, res)
}

func (c *localCommContr) CommentsResListener(ctx context.Context, name string) error {
	return c.d.resListener(ctx, name, messaging.CommentsSetup)
}

func (c *localCommContr) Name() string {
	return c.d.name
}
