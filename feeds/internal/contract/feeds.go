package contract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/booleanism/tetek/feeds/infra/amqp"
	"github.com/booleanism/tetek/feeds/internal/entities"
	"github.com/booleanism/tetek/feeds/internal/model/pools"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"github.com/rabbitmq/amqp091-go"
)

type FeedsContr struct {
	con  *amqp091.Connection
	repo repo.FeedsRepo
}

func NewFeeds(con *amqp091.Connection, repo repo.FeedsRepo) *FeedsContr {
	return &FeedsContr{con, repo}
}

func (c *FeedsContr) WorkerFeedsListener(ctx context.Context) (*amqp091.Channel, error) {
	ctx, log := loggr.GetLogger(ctx, "worker")
	ch, err := c.con.Channel()
	if err != nil {
		log.Error(err, "failed to create channel")
		return nil, err
	}

	err = amqp.FeedsSetup(ch)
	if err != nil {
		log.Error(err, "failed to setup feeds topic")
		return nil, err
	}

	mgs, err := ch.Consume(amqp.FeedsTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Error(err, "failed to consume feeds task")
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

			task := amqp.FeedsTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.V(1).Info("failed to marshal feeds task", "requestID", d.CorrelationId, "error", err, "body", d.Body)
				res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsParseFail, Message: "error parsing feeds task"})
				feedsResultPublisher(log, task, ch, d, res, err)
				continue
			}

			func() {
				fBuf, ok := pools.FeedsPool.Get().(*pools.Feeds)
				if !ok {
					// TODO: logging
					if err := d.Nack(false, false); err != nil {
						return
					}
					return
				}
				defer pools.FeedsPool.Put(fBuf)
				defer fBuf.Reset()

				if task.Cmd == 0 {
					ff := repo.FeedsFilter{ID: task.ID}
					err := c.repo.Feeds(ctx, ff, &fBuf.Value)
					if err == nil {
						res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.Success, Message: "feeds found", Detail: fBuf.Value[len(fBuf.Value)-1]})
						feedsResultPublisher(log, task, ch, d, res, err)
						return
					}

					if err == pgx.ErrNoRows {
						res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsNoFeeds, Message: "feeds not found", Detail: entities.Feed{ID: ff.ID}})
						feedsResultPublisher(log, task, ch, d, res, err)
						return
					}

					res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsDBError, Message: "something happen in our end", Detail: entities.Feed{ID: ff.ID}})
					feedsResultPublisher(log, task, ch, d, res, err)
					return
				}

				res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsUnknownCmd, Message: "unknown command"})
				feedsResultPublisher(log, task, ch, d, res, fmt.Errorf("unexpected command"))
			}()
		}
	}()

	return ch, nil
}

func feedsResultPublisher(log logr.Logger, task any, ch *amqp091.Channel, d amqp091.Delivery, res []byte, e error) {
	log = log.WithName("feeds-result-publisher")
	err := ch.Publish(amqp.FeedsExchange, amqp.FeedsResRk, false, false, amqp091.Publishing{
		CorrelationId: d.CorrelationId,
		Body:          res,
		ContentType:   "text/json",
	})
	if err != nil {
		log.Error(err, "failed to publish feeds result", "result", res, "exchange", amqp.FeedsExchange, "routing-key", amqp.FeedsResRk)
		helper.Nack(log, d, "unable to publish feeds result", "task", task)
		return
	}

	if e != nil {
		helper.Nack(log, d, e.Error(), "task", task)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Error(err, "failed to ack feeds task", "task", task)
		return
	}
}
