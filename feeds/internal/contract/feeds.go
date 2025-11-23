package contract

import (
	"context"
	"encoding/json"

	"github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/pools"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
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

func (c *FeedsContr) WorkerFeedsListener() (*amqp091.Channel, error) {
	ch, err := c.con.Channel()
	if err != nil {
		return nil, err
	}

	err = amqp.FeedsSetup(ch)
	if err != nil {
		return nil, err
	}

	mgs, err := ch.Consume(amqp.FeedsTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range mgs {
			if d.ContentType != "text/json" {
				continue
			}

			task := amqp.FeedsTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsParseFail, Message: "error parsing feeds task"})
				if err := ch.Publish(amqp.FeedsExchange, amqp.FeedsResRk, false, false, amqp091.Publishing{
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
				func() {
					ff := repo.FeedsFilter{ID: task.ID}
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

					err := c.repo.Feeds(context.Background(), ff, &fBuf.Value)
					if err == nil {
						res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.Success, Message: "feeds found", Detail: fBuf.Value[len(fBuf.Value)-1]})
						if err := ch.Publish(amqp.FeedsExchange, amqp.FeedsResRk, false, false, amqp091.Publishing{
							CorrelationId: d.CorrelationId,
							Body:          res,
							ContentType:   "text/json",
						}); err != nil {
							return
						}
						if err := d.Ack(false); err != nil {
							return
						}
						return
					}

					if err == pgx.ErrNoRows {
						res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsNoFeeds, Message: "feeds not found", Detail: model.Feed{ID: ff.ID}})
						if err := ch.Publish(amqp.FeedsExchange, amqp.FeedsResRk, false, false, amqp091.Publishing{
							CorrelationId: d.CorrelationId,
							Body:          res,
							ContentType:   "text/json",
						}); err != nil {
							return
						}
						if err := d.Ack(false); err != nil {
							return
						}
						return
					}

					res, _ := json.Marshal(&amqp.FeedsResult{Code: errro.ErrFeedsDBError, Message: "something happen in our end", Detail: model.Feed{ID: ff.ID}})
					if err := ch.Publish(amqp.FeedsExchange, amqp.FeedsResRk, false, false, amqp091.Publishing{
						CorrelationId: d.CorrelationId,
						Body:          res,
						ContentType:   "text/json",
					}); err != nil {
						return
					}
					if err := d.Ack(false); err != nil {
						return
					}
				}()
			}
		}
	}()

	return ch, nil
}
