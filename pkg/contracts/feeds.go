package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/rabbitmq/amqp091-go"
)

type FeedsSubsribe interface {
	Publish(context.Context, amqp.FeedsTask) error
	Consume(context.Context, **amqp.FeedsResult) error
}

type localFeedsContr struct {
	con  *amqp091.Connection
	res  map[string]chan *amqp.FeedsResult
	mRes sync.Mutex
}

func SubscribeFeeds(con *amqp091.Connection, subcriberName string) *localFeedsContr {
	self := &localFeedsContr{con: con, res: make(map[string]chan *amqp.FeedsResult)}
	if err := self.feedsResListener(subcriberName); err != nil {
		panic(err)
	}

	return self
}

func (c *localFeedsContr) Publish(ctx context.Context, task amqp.FeedsTask) error {
	corrID := ctx.Value(keystore.RequestId{}).(string)
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
	c.res[corrID] = make(chan *amqp.FeedsResult)
	c.mRes.Unlock()

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	if err = ch.Publish(amqp.FeedsExchange, amqp.FeedsTaskRk, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		return err
	}

	return nil
}

func (c *localFeedsContr) Consume(ctx context.Context, res **amqp.FeedsResult) error {
	corrID := ctx.Value(keystore.RequestId{}).(string)
	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		return errors.New("no result with given correlation id")
	}

	select {
	case <-ctx.Done():
		return errors.New("failed to consume feeds task")
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrID)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *localFeedsContr) feedsResListener(name string) error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.FeedsSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(fmt.Sprintf("feeds_res_%s", name), true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, amqp.FeedsResRk, amqp.FeedsExchange, false, nil)
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

			var res amqp.FeedsResult
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
