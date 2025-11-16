package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/rabbitmq/amqp091-go"
)

type AuthSubscribe interface {
	Publish(context.Context, amqp.AuthTask) error
	Consume(context.Context, **amqp.AuthResult) error
}

type localAuthContr struct {
	con  *amqp091.Connection
	res  map[string]chan *amqp.AuthResult
	mRes sync.Mutex
}

func SubsribeAuth(con *amqp091.Connection, subcriberName string) *localAuthContr {
	self := &localAuthContr{con: con, res: make(map[string]chan *amqp.AuthResult)}
	if err := self.authResListener(subcriberName); err != nil {
		panic(err)
	}

	return self
}

func (c *localAuthContr) Publish(ctx context.Context, task amqp.AuthTask) error {
	corrId := ctx.Value(keystore.RequestId{}).(string)
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	c.mRes.Lock()
	c.res[corrId] = make(chan *amqp.AuthResult, 1)
	c.mRes.Unlock()

	if err = ch.Publish(amqp.AUTH_EXCHANGE, amqp.AUTH_TASK_RK, false, false, amqp091.Publishing{
		CorrelationId: corrId,
		Body:          t,
		ContentType:   "text/json",
	}); err != nil {
		return err
	}

	return nil
}

func (c *localAuthContr) Consume(ctx context.Context, res **amqp.AuthResult) error {
	corrId := ctx.Value(keystore.RequestId{}).(string)
	c.mRes.Lock()
	ch, ok := c.res[corrId]
	c.mRes.Unlock()
	if !ok {
		return errors.New("no result with given correlation id")
	}

	select {
	case <-ctx.Done():
		return errors.New("failed to consume auth result")
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrId)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *localAuthContr) authResListener(name string) error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.AuthSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(fmt.Sprintf("auth_res_%s", name), true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, amqp.AUTH_RES_RK, amqp.AUTH_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	mgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range mgs {
			c.mRes.Lock()
			respCh, ok := c.res[d.CorrelationId]
			c.mRes.Unlock()

			if !ok {
				d.Nack(false, false)
				continue
			}

			if d.ContentType != "text/json" {
				continue
			}

			var res = amqp.AuthResult{}
			if err := json.Unmarshal(d.Body, &res); err != nil {
				d.Nack(false, false)
				continue
			}

			select {
			case respCh <- &res:
				d.Ack(false)
			default:
				d.Nack(false, false)
			}
		}
	}()

	return nil
}
