package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
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

func SubscribeAccount(con *amqp091.Connection) *localAccContr {
	return &localAccContr{con: con, res: make(map[string]chan *amqp.AccountRes)}
}

func (c *localAccContr) Publish(ctx context.Context, task amqp.AccountTask) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, "account-task-publisher")

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
	c.res[corrID] = make(chan *amqp.AccountRes)
	c.mRes.Unlock()

	t, _ := json.Marshal(&task)
	if err = ch.Publish(amqp.AccountExchange, amqp.AccountTaskRk, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		log.Error(err, "failed to publish account task", "task", task)
		return err
	}

	return nil
}

func (c *localAccContr) Consume(ctx context.Context, res **amqp.AccountRes) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, "account-task-publisher")

	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		e := errors.New("no account result with given correlation id")
		log.V(1).Info("does not receive account result from chan", "error", e)
		return e
	}

	select {
	case <-ctx.Done():
		e := errors.New("deadline exceeded")
		log.V(1).Info("receive ctx.Done", "error", e)
		return e
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrID)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *localAccContr) AccountResListener(ctx context.Context, name string) error {
	_, log := loggr.GetLogger(ctx, "account-res-listener")
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

	err = ch.QueueBind(q.Name, amqp.AccountResRk, amqp.AccountExchange, false, nil)
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

			res := amqp.AccountRes{}
			if err := json.Unmarshal(d.Body, &res); err != nil {
				log.V(1).Info("failed to marshal account result", "requestID", d.CorrelationId, "body", d.Body)
				helper.Nack(log, d, "failed to marshal account result", "requestID", d.CorrelationId, "body", d.Body)
				continue
			}

			select {
			case respCh <- &res:
				if err := d.Ack(false); err != nil {
					log.Error(err, "failed to nack account result", "requestID", d.CorrelationId, "message", d)
					continue
				}
			default:
				helper.Nack(log, d, "cannot sent auth result into result chan", "requestID", d.CorrelationId, "result", res)
				log.V(1).Info("cannot sent auth result into result chan", "requestID", d.CorrelationId, "result", res)
				continue
			}
		}
	}()

	return nil
}
