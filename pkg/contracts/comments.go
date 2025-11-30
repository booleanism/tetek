package contracts

import (
	"context"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/rabbitmq/amqp091-go"
)

type CommentsDealer interface {
	Dealer
}

type localCommContr struct {
	d DealContract[amqp.CommentsTask, amqp.CommentsResult]
}

func CommentsAssent(con *amqp091.Connection) *localCommContr {
	return &localCommContr{
		d: DealContract[amqp.CommentsTask, amqp.CommentsResult]{
			con:      con,
			name:     "comments",
			exchange: amqp.CommentsExchange,
			trk:      amqp.CommentsTaskRk,
			rrk:      amqp.CommentsResRk,
			res:      make(map[string]chan *amqp.CommentsResult),
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
	return c.d.resListener(ctx, name, amqp.CommentsSetup)
}

func (c *localCommContr) Name() string {
	return c.d.name
}
