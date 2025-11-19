package contract

import (
	"context"
	"encoding/json"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
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
	_, log := loggr.GetLogger(ctx, "worker")
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
			log = log.WithValues("requestID", d.CorrelationId)
			if d.ContentType != "text/json" {
				log.V(1).Info("unexpected ContentType", "ContentType", d.ContentType)
				if err := d.Nack(false, false); err != nil {
					log.Error(err, "failed to nack account task", "body", d.Body, "ContentType", d.ContentType)
				}
				continue
			}

			task := amqp.AccountTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.V(1).Info("failed to marshal account task", "error", err, "body", d.Body)
				res, _ := json.Marshal(&amqp.AccountRes{Code: errro.ErrAccountParseFail, Message: "error parsing account task"})
				resultPublisher(log, task, ch, d, res)
				continue
			}

			if task.Cmd == 0 {
				u, err := c.repo.GetUser(context.Background(), task.User)
				if err == nil {
					log.V(2).Info("user found", "task", task)
					res, _ := json.Marshal(&amqp.AccountRes{Code: errro.Success, Message: "user found", Detail: u})
					resultPublisher(log, task, ch, d, res)
					continue
				}

				if err == pgx.ErrNoRows {
					log.V(2).Info("no user", "error", err, "task", task)
					res, _ := json.Marshal(&amqp.AccountRes{Code: errro.ErrAccountNoUser, Message: "user not found", Detail: task.User})
					resultPublisher(log, task, ch, d, res)
					continue
				}

				log.Error(err, "unexpected error", "task", task)
				res, _ := json.Marshal(&amqp.AccountRes{Code: errro.ErrAccountDBError, Message: "something happen in our end", Detail: task.User})
				resultPublisher(log, task, ch, d, res)
				continue
			}
		}
	}()

	return ch, nil
}

func resultPublisher(log logr.Logger, task amqp.AccountTask, ch *amqp091.Channel, d amqp091.Delivery, res []byte) {
	log = log.WithName("result-publisher")
	err := ch.Publish(amqp.AccountExchange, amqp.AccountResRk, false, false, amqp091.Publishing{
		CorrelationId: d.CorrelationId,
		Body:          res,
		ContentType:   "text/json",
	})
	if err != nil {
		log.Error(err, "failed to publish account result", "result", res, "exchange", amqp.AccountExchange, "routing-key", amqp.AccountResRk)
		if err := d.Nack(false, false); err != nil {
			log.Error(err, "failed to nack account task", "task", task)
		}
		return
	}

	if err := d.Ack(false); err != nil {
		log.Error(err, "failed to ack account task", "task", task)
		return
	}
}
