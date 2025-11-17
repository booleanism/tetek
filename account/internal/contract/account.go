package contract

import (
	"context"
	"encoding/json"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/jackc/pgx/v5"
	"github.com/rabbitmq/amqp091-go"
)

type AccContr struct {
	con  *amqp091.Connection
	repo repo.UserRepo
}

func NewAccount(con *amqp091.Connection, repo repo.UserRepo) *AccContr {
	return &AccContr{con, repo}
}

func (c *AccContr) WorkerAccountListener() (*amqp091.Channel, error) {
	ch, err := c.con.Channel()
	if err != nil {
		return nil, err
	}

	err = amqp.AccountSetup(ch)
	if err != nil {
		return nil, err
	}

	mgs, err := ch.Consume(amqp.ACCOUNT_TASK_QUEUE, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range mgs {
			if d.ContentType != "text/json" {
				continue
			}

			task := amqp.AccountTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				res, _ := json.Marshal(&amqp.AccountRes{Code: errro.EACCOUNT_PARSE_FAIL, Message: "error parsing account task"})
				if err := ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_RES_RK, false, false, amqp091.Publishing{
					CorrelationId: d.CorrelationId,
					Body:          res,
					ContentType:   "text/json",
				}); err != nil {
					continue
				}
				if err := d.Nack(false, false); err != nil {
					continue
				}
				continue
			}

			if task.Cmd == 0 {
				u, err := c.repo.GetUser(context.Background(), task.User)
				if err == nil {
					res, _ := json.Marshal(&amqp.AccountRes{Code: errro.SUCCESS, Message: "user found", Detail: u})
					if err := ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_RES_RK, false, false, amqp091.Publishing{
						CorrelationId: d.CorrelationId,
						Body:          res,
						ContentType:   "text/json",
					}); err != nil {
						continue
					}
					if err := d.Ack(false); err != nil {
						continue
					}
					continue
				}

				if err == pgx.ErrNoRows {
					res, _ := json.Marshal(&amqp.AccountRes{Code: errro.EACCOUNT_NO_USER, Message: "user not found", Detail: task.User})
					if err := ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_RES_RK, false, false, amqp091.Publishing{
						CorrelationId: d.CorrelationId,
						Body:          res,
						ContentType:   "text/json",
					}); err != nil {
						continue
					}
					if err := d.Ack(false); err != nil {
						continue
					}
					continue
				}

				res, _ := json.Marshal(&amqp.AccountRes{Code: errro.EACCOUNT_DB_ERR, Message: "something happen in our end", Detail: task.User})
				if err := ch.Publish(amqp.ACCOUNT_EXCHANGE, amqp.ACCOUNT_RES_RK, false, false, amqp091.Publishing{
					CorrelationId: d.CorrelationId,
					Body:          res,
					ContentType:   "text/json",
				}); err != nil {
					continue
				}
				if err := d.Ack(false); err != nil {
					continue
				}
			}
		}
	}()

	return ch, nil
}
