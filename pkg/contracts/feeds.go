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
	corrId := ctx.Value(keystore.RequestId{}).(string)
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	c.mRes.Lock()
	c.res[corrId] = make(chan *amqp.FeedsResult)
	c.mRes.Unlock()

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	if err = ch.Publish(amqp.FEEDS_EXCHANGE, amqp.FEEDS_TASK_RK, false, false, amqp091.Publishing{
		CorrelationId: corrId,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		return err
	}

	return nil
}

func (c *localFeedsContr) Consume(ctx context.Context, res **amqp.FeedsResult) error {
	corrId := ctx.Value(keystore.RequestId{}).(string)
	c.mRes.Lock()
	ch, ok := c.res[corrId]
	c.mRes.Unlock()
	if !ok {
		return errors.New("no result with given correlation id")
	}

	select {
	case <-ctx.Done():
		return errors.New("failed to consume feeds task")
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrId)
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

	err = ch.QueueBind(q.Name, amqp.FEEDS_RES_RK, amqp.FEEDS_EXCHANGE, false, nil)
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

			var res amqp.FeedsResult
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
