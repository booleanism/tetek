package contract

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/booleanism/tetek/auth/amqp"
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

func (c *LocalAuthContr) Publish(corrID string, task amqp.AuthTask) error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
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
		return err
	}

	return nil
}

func (c *LocalAuthContr) Consume(corrID string) (*amqp.AuthResult, error) {
	c.mRes.Lock()
	ch, ok := c.res[corrID]
	c.mRes.Unlock()
	if !ok {
		return nil, errors.New("no result with given correlation id")
	}

	res := <-ch

	c.mRes.Lock()
	delete(c.res, corrID)
	close(ch)
	c.mRes.Unlock()

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

			res := amqp.AuthResult{}
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
