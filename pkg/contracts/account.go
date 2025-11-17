package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/rabbitmq/amqp091-go"
)

type AccountSubscribe interface {
	Publish(context.Context, amqp.AccountTask) error
	Consume(context.Context, **amqp.AccountRes) error
}

type localAccContr struct {
	con  *amqp091.Connection
	res  map[string]chan *amqp.AccountRes
	mRes sync.Mutex
}

func SubscribeAccount(con *amqp091.Connection, subcriberName string) *localAccContr {
	self := &localAccContr{con: con, res: make(map[string]chan *amqp.AccountRes)}
	if err := self.accountResListener(subcriberName); err != nil {
		panic(err)
	}

	return self
}

func (c *localAccContr) Publish(ctx context.Context, task amqp.AccountTask) error {
	corrID, ok := ctx.Value(keystore.RequestId{}).(string)
	if !ok {
		return errors.New("no requestId")
	}

	ch, err := c.con.Channel()
	if err != nil {
		return err
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	c.mRes.Lock()
	c.res[corrID] = make(chan *amqp.AccountRes)
	c.mRes.Unlock()

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	if err = ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_TASK_RK, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		return err
	}

	return nil
}

func (c *localAccContr) Consume(ctx context.Context, res **amqp.AccountRes) error {
	corrID := ctx.Value(keystore.RequestId{}).(string)
	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		return errors.New("no result with given correlation id")
	}

	select {
	case <-ctx.Done():
		return errors.New("failed to consume account task")
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrID)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *localAccContr) accountResListener(name string) error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.AccountSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(fmt.Sprintf("account_res_%s", name), true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, amqp.ACCOUNT_RES_RK, amqp.ACCOUNT_EXCHANGE, false, nil)
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
				if err := d.Nack(false, false); err != nil {
					continue
				}
				continue
			}

			if d.ContentType != "text/json" {
				continue
			}

			var res amqp.AccountRes
			if err := json.Unmarshal(d.Body, &res); err != nil {
				if err := d.Nack(false, false); err != nil {
					continue
				}
				continue
			}

			select {
			case respCh <- &res:
				if err := d.Ack(false); err != nil {
					continue
				}
			default:
				if err := d.Nack(false, false); err != nil {
					continue
				}
			}
		}
	}()

	return nil
}
