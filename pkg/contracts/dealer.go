package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/rabbitmq/amqp091-go"
)

type Dealer interface {
	Publish(ctx context.Context, task any) error
	Consume(ctx context.Context, result any) error
	Name() string
}

type DealContract[T any, R any] struct {
	name     string
	exchange string
	trk      string
	rrk      string
	con      *amqp091.Connection
	res      map[string]chan *R
	mRes     sync.Mutex
}

func (c *DealContract[_, R]) publish(ctx context.Context, task any) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, fmt.Sprintf("%s-task-publisher", c.name))

	ch, err := c.con.Channel()
	if err != nil {
		log.Error(err, "failed to open channel")
		return err
	}
	defer func() {
		if err := ch.Close(); err != nil {
			log.Error(err, "failed to close channel")
		}
	}()

	c.mRes.Lock()
	c.res[corrID] = make(chan *R)
	c.mRes.Unlock()

	t, _ := json.Marshal(&task)
	if err = ch.Publish(c.exchange, c.trk, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		log.Error(err, fmt.Sprintf("failed to publish %s task", c.exchange), "task", task)
		return err
	}

	return nil
}

func (c *DealContract[_, R]) consume(ctx context.Context, res any) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, fmt.Sprintf("%s-result-consumer", c.name))

	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		e := fmt.Errorf("no %s with given correlation id", c.name)
		log.V(1).Info(fmt.Sprintf("does not receive %s result from chan", c.name), "error", e)
		return e
	}

	select {
	case <-ctx.Done():
		e := errors.New("deadline exceeded")
		log.V(1).Info("receive ctx.Done", "error", e)
		return e
	case *res.(**R) = <-ch:
		c.mRes.Lock()
		delete(c.res, corrID)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *DealContract[T, R]) resListener(ctx context.Context, name string, setup func(ch *amqp091.Channel) error) error {
	_, log := loggr.GetLogger(ctx, fmt.Sprintf("%s-res-listener", c.name))
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := setup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(fmt.Sprintf("account_res_%s", name), true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, c.rrk, c.exchange, false, nil)
	if err != nil {
		return err
	}

	mgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range mgs {
			if d.ContentType != "text/json" {
				log.V(1).Info("unexpected ContentType", "requestID", d.CorrelationId, "ContentType", d.ContentType)
				helper.Nack(log, d, "unexpected ContentType")
				continue
			}

			c.mRes.Lock()
			respCh, ok := c.res[d.CorrelationId]
			c.mRes.Unlock()
			if !ok {
				helper.Nack(log, d, "requestID", d.CorrelationId, "res", c.res)
				continue
			}

			var res R
			if err := json.Unmarshal(d.Body, &res); err != nil {
				log.V(1).Info(fmt.Sprintf("failed to marshal %s result", c.name), "requestID", d.CorrelationId, "body", d.Body)
				helper.Nack(log, d, fmt.Sprintf("failed to marshal %s result", c.name), "requestID", d.CorrelationId, "body", d.Body)
				continue
			}

			select {
			case respCh <- &res:
				if err := d.Ack(false); err != nil {
					log.Error(err, fmt.Sprintf("failed to nack %s result", c.name), "requestID", d.CorrelationId, "message", d)
					continue
				}
			default:
				helper.Nack(log, d, fmt.Sprintf("cannot sent %s result into result chan", c.name), "requestID", d.CorrelationId, "result", res)
				log.V(1).Info(fmt.Sprintf("cannot sent %s result into result chan", c.name), "requestID", d.CorrelationId, "result", res)
				continue
			}
		}
	}()

	return nil
}
