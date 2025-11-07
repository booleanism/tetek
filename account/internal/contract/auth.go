package contract

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/rabbitmq/amqp091-go"
)

type LocalAuthContr struct {
	con  *amqp091.Connection
	res  map[string]chan *amqp.AuthResult
	mRes sync.Mutex
}

func NewAuth(con *amqp091.Connection) *LocalAuthContr {
	self := &LocalAuthContr{con: con, res: make(map[string]chan *amqp.AuthResult)}
	if err := self.authResListener(); err != nil {
		panic(err)
	}

	return self
}

func (c *LocalAuthContr) Publish(corrId string, task amqp.AuthTask) error {
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
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("auth task sent", "id", corrId, "task", task)
	})

	return nil
}

func (c *LocalAuthContr) Consume(corrId string) (*amqp.AuthResult, error) {
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
		z.V(4).Info("auth result eaten", "id", corrId)
	})

	return res, nil
}

func (c *LocalAuthContr) authResListener() error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.AuthSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare("auth_res_account", true, false, false, false, nil)
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
			loggr.LogInfo(func(z loggr.LogInf) {
				z.V(4).Info("receive new auth result", "id", d.CorrelationId, "body", d.Body)
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

			var res = amqp.AuthResult{}
			if err := json.Unmarshal(d.Body, &res); err != nil {
				loggr.LogError(func(z loggr.LogErr) errro.Error {
					z.V(0).Error(err, "parsing auth result failed", "id", d.CorrelationId)
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
