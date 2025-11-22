package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
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

func SubsribeAuth(con *amqp091.Connection) *localAuthContr {
	return &localAuthContr{con: con, res: make(map[string]chan *amqp.AuthResult)}
}

func (c *localAuthContr) Publish(ctx context.Context, task amqp.AuthTask) error {
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

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	c.mRes.Lock()
	c.res[corrID] = make(chan *amqp.AuthResult, 1)
	c.mRes.Unlock()

	if err = ch.Publish(amqp.AuthExchange, amqp.AuthTaskRk, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		Body:          t,
		ContentType:   "text/json",
	}); err != nil {
		log.Error(err, "failed to publish auth task", "task", task)
		return err
	}

	return nil
}

func (c *localAuthContr) Consume(ctx context.Context, res **amqp.AuthResult) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, "auth-task-publisher")

	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		e := errors.New("no auth result with given correlation id")
		log.V(1).Info("does not receive account result from chan", "id", corrID, "error", e)
		return e
	}

	select {
	case <-ctx.Done():
		e := errors.New("failed to consume auth result")
		log.V(1).Info("receive ctx.Done", "id", corrID, "error", e)
		return e
	case *res = <-ch:
		c.mRes.Lock()
		delete(c.res, corrID)
		close(ch)
		c.mRes.Unlock()

		return nil
	}
}

func (c *localAuthContr) AuthResListener(ctx context.Context, name string) error {
	_, log := loggr.GetLogger(ctx, "auth-res-listener")
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

	err = ch.QueueBind(q.Name, amqp.AuthResRk, amqp.AuthExchange, false, nil)
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

			res := amqp.AuthResult{}
			if err := json.Unmarshal(d.Body, &res); err != nil {
				log.V(1).Info("failed to marshal auth result", "requestID", d.CorrelationId, "body", d.Body)
				helper.Nack(log, d, "failed to marshal auth result", "requestID", d.CorrelationId, "body", d.Body)
				continue
			}

			select {
			case respCh <- &res:
				if err := d.Ack(false); err != nil {
					log.Error(err, "failed to nack the auth result", "requestID", d.CorrelationId, "message", d)
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
