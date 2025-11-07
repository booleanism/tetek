package contract

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/rabbitmq/amqp091-go"
)

type LocalAccContr struct {
	con  *amqp091.Connection
	res  map[string]chan *amqp.AccountRes
	mRes sync.Mutex
}

func NewAccount(con *amqp091.Connection) *LocalAccContr {
	self := &LocalAccContr{con: con, res: make(map[string]chan *amqp.AccountRes)}
	if err := self.accountResListener(); err != nil {
		panic(err)
	}

	return self
}

func (c *LocalAccContr) Publish(corrId string, task amqp.AccountTask) error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	c.mRes.Lock()
	c.res[corrId] = make(chan *amqp.AccountRes)
	c.mRes.Unlock()

	t, err := json.Marshal(&task)
	if err != nil {
		return err
	}

	if err = ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_TASK_RK, false, false, amqp091.Publishing{
		CorrelationId: corrId,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		return err
	}
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("account task sent", "id", corrId, "task", task)
	})

	return nil
}

func (c *LocalAccContr) Consume(corrId string) (*amqp.AccountRes, error) {
	c.mRes.Lock()
	ch, ok := c.res[corrId]
	c.mRes.Unlock()
	if !ok {
		return nil, errors.New("no result with given correlation id")
	}

	res := <-ch

	c.mRes.Lock()
	delete(c.res, corrId)
	close(ch)
	c.mRes.Unlock()
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("account result eaten", "id", corrId)
	})

	return res, nil
}

func (c *LocalAccContr) accountResListener() error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.AccountSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare("account_res_feeds", true, false, false, false, nil)
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
			loggr.LogInfo(func(z loggr.LogInf) {
				z.V(4).Info("receive new account result", "id", d.CorrelationId, "body", d.Body)
			})
			c.mRes.Lock()
			respCh, ok := c.res[d.CorrelationId]
			c.mRes.Unlock()

			if !ok {
				loggr.LogInfo(func(z loggr.LogInf) {
					z.V(4).Info("no waiting receiver for corrId, skipping", "id", d.CorrelationId)
				})
				d.Nack(false, false)
				continue
			}

			if d.ContentType != "text/json" {
				loggr.LogInfo(func(z loggr.LogInf) {
					z.V(4).Info("content type missmatch, skipping", "id", d.CorrelationId)
				})
				continue
			}

			var res amqp.AccountRes
			if err := json.Unmarshal(d.Body, &res); err != nil {
				loggr.LogError(func(z loggr.LogErr) errro.Error {
					z.V(0).Error(err, "parsing account result failed", "d", d.CorrelationId)
					return nil
				})
				d.Nack(false, false)
				continue
			}

			select {
			case respCh <- &res:
				d.Ack(false)
			default:
				loggr.LogInfo(func(z loggr.LogInf) {
					z.V(3).Info("receiver not ready, dropping", "id", d.CorrelationId)
				})
				d.Nack(false, false)
			}
		}
	}()

	return nil
}
