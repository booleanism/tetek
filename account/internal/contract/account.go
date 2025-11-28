package contract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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

func (c *AccContr) WorkerAccountListener(ctx context.Context) (*amqp091.Channel, error) {
	ctx, log := loggr.GetLogger(ctx, "worker")
	ch, err := c.con.Channel()
	if err != nil {
		log.Error(err, "failed to create channel")
		return nil, err
	}

	err = amqp.AccountSetup(ch)
	if err != nil {
		log.Error(err, "failed to setup account topic")
		return nil, err
	}

	mgs, err := ch.Consume(amqp.AccountTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Error(err, "failed to consume account task")
		return nil, err
	}

	go func() {
		for d := range mgs {
			ctx = context.WithValue(ctx, keystore.RequestID{}, d.CorrelationId)
			if d.ContentType != "text/json" {
				log.V(1).Info("unexpected ContentType", "ContentType", d.ContentType)
				helper.Nack(log, d, "unexpected ContentType", "requestID", d.CorrelationId)
				continue
			}

			task := amqp.AccountTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.V(1).Info("failed to marshal account task", "requestID", d.CorrelationId, "error", err, "body", d.Body)
				res, _ := json.Marshal(&amqp.AccountResult{Code: errro.ErrAccountParseFail, Message: "error parsing account task"})
				accountResultPublisher(log, task, ch, d, res, err)
				continue
			}

			if task.Cmd == 0 {
				u := &task.User
				err := c.repo.GetUser(ctx, &u)
				if err == nil {
					res, _ := json.Marshal(&amqp.AccountResult{Code: errro.Success, Message: "user found", Detail: *u})
					accountResultPublisher(log, task, ch, d, res, err)
					continue
				}

				if err == pgx.ErrNoRows {
					res, _ := json.Marshal(&amqp.AccountResult{Code: errro.ErrAccountNoUser, Message: "user not found", Detail: task.User})
					accountResultPublisher(log, task, ch, d, res, err)
					continue
				}

				res, _ := json.Marshal(&amqp.AccountResult{Code: errro.ErrAccountDBError, Message: "something happen in our end", Detail: task.User})
				accountResultPublisher(log, task, ch, d, res, err)
				continue
			}

			res, _ := json.Marshal(&amqp.AccountResult{Code: errro.ErrCommUnknownCmd, Message: "unknown command"})
			accountResultPublisher(log, task, ch, d, res, fmt.Errorf("unexpected command"))
		}
	}()

	return ch, nil
}

func accountResultPublisher(log logr.Logger, task any, ch *amqp091.Channel, d amqp091.Delivery, res []byte, e error) {
	log = log.WithName("account-result-publisher")
	err := ch.Publish(amqp.AccountExchange, amqp.AccountResRk, false, false, amqp091.Publishing{
		CorrelationId: d.CorrelationId,
		Body:          res,
		ContentType:   "text/json",
	})
	if err != nil {
		log.Error(err, "failed to publish account result", "result", res, "exchange", amqp.AccountExchange, "routing-key", amqp.AccountResRk)
		helper.Nack(log, d, "unable to publish account result", "task", task)
		return
	}

	if e != nil {
		helper.Nack(log, d, e.Error(), "task", task)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Error(err, "failed to ack account task", "task", task)
		return
	}
}
