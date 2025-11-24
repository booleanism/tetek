package contract

import (
	"context"
	"encoding/json"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/comments/internal/pools"
	"github.com/booleanism/tetek/comments/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/jackc/pgx/v5"
	"github.com/rabbitmq/amqp091-go"
)

type CommentsContr struct {
	con  *amqp091.Connection
	repo repo.CommentsRepo
}

func NewComments(con *amqp091.Connection, repo repo.CommentsRepo) *CommentsContr {
	return &CommentsContr{con, repo}
}

func (c *CommentsContr) WorkerCommentsListener() (*amqp091.Channel, error) {
	ch, err := c.con.Channel()
	if err != nil {
		return nil, err
	}

	err = amqp.CommentsSetup(ch)
	if err != nil {
		return nil, err
	}

	mgs, err := ch.Consume(amqp.CommentsTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range mgs {
			if d.ContentType != "text/json" {
				continue
			}

			task := amqp.CommentsTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommParseFail, Message: "error parsing comments task"})
				if err := ch.Publish(amqp.CommentsExchange, amqp.CommentsResRk, false, false, amqp091.Publishing{
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
					cf := repo.CommentFilter{Head: task.Parent}
					fBuf, ok := pools.CommentsPool.Get().(*pools.Comments)
					if !ok {
						// TODO: logging
						if err := d.Nack(false, false); err != nil {
							return
						}
						return
					}
					defer pools.CommentsPool.Put(fBuf)
					defer fBuf.Reset()

					n, err := c.repo.GetComments(context.Background(), cf, &fBuf.Value)
					if err == nil || n != 0 {
						res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.Success, Message: "comments found", Details: fBuf.Value[0:n]})
						if err := ch.Publish(amqp.CommentsExchange, amqp.CommentsResRk, false, false, amqp091.Publishing{
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
						res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommNoComments, Message: "comments not found"})
						if err := ch.Publish(amqp.CommentsExchange, amqp.CommentsResRk, false, false, amqp091.Publishing{
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

					res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommDBError, Message: "something happen in our end"})
					if err := ch.Publish(amqp.CommentsExchange, amqp.CommentsResRk, false, false, amqp091.Publishing{
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
