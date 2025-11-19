package contract

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/booleanism/tetek/account/amqp"
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

func (c *LocalAccContr) Publish(corrID string, task amqp.AccountTask) error {
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

	if err = ch.Publish(amqp.AccountExchange, amqp.AccountTaskRk, false, false, amqp091.Publishing{
		CorrelationId: corrID,
		ContentType:   "text/json",
		Body:          t,
	}); err != nil {
		return err
	}

	return nil
}

func (c *LocalAccContr) Consume(corrID string) (*amqp.AccountRes, error) {
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

func (c *LocalAccContr) accountResListener() error {
	ch, err := c.con.Channel()
	if err != nil {
		return err
	}

	if err := amqp.AccountSetup(ch); err != nil {
		return err
	}

	q, err := ch.QueueDeclare("account_res_auth", true, false, false, false, nil)
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
