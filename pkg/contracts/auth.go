package contracts

import (
	"context"

	messaging "github.com/booleanism/tetek/auth/infra/messaging/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type AuthDealer interface {
	Dealer
}

type localAuthContr struct {
	d DealContract[messaging.AuthTask, messaging.AuthResult]
}

func AuthAssent(con *amqp091.Connection) *localAuthContr {
	return &localAuthContr{d: DealContract[messaging.AuthTask, messaging.AuthResult]{
		con:      con,
		name:     "auth",
		exchange: messaging.AuthExchange,
		trk:      messaging.AuthTaskRk,
		rrk:      messaging.AuthResRk,
		res:      make(map[string]chan *messaging.AuthResult),
	}}
}

func (c *localAuthContr) Publish(ctx context.Context, task any) error {
	return c.d.publish(ctx, task)
}

func (c *localAuthContr) Consume(ctx context.Context, res any) error {
	return c.d.consume(ctx, res)
}

func (c *localAuthContr) AuthResListener(ctx context.Context, name string) error {
	return c.d.resListener(ctx, name, messaging.AuthSetup)
}

func (c *localAuthContr) Name() string {
	return c.d.name
}
