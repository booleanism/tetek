package contracts

import (
	"context"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/rabbitmq/amqp091-go"
)

type AuthDealer interface {
	Dealer
}

type localAuthContr struct {
	d DealContract[amqp.AuthTask, amqp.AuthResult]
}

func AuthAssent(con *amqp091.Connection) *localAuthContr {
	return &localAuthContr{d: DealContract[amqp.AuthTask, amqp.AuthResult]{
		con:      con,
		name:     "auth",
		exchange: amqp.AuthExchange,
		trk:      amqp.AuthTaskRk,
		rrk:      amqp.AuthResRk,
		res:      make(map[string]chan *amqp.AuthResult),
	}}
}

func (c *localAuthContr) Publish(ctx context.Context, task any) error {
	return c.d.Publish(ctx, task)
}

func (c *localAuthContr) Consume(ctx context.Context, res any) error {
	return c.d.Consume(ctx, res)
}

func (c *localAuthContr) AuthResListener(ctx context.Context, name string) error {
	return c.d.resListener(ctx, name, amqp.AuthSetup)
}

func (c *localAuthContr) Name() string {
	return c.d.name
}
