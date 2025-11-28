package contract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/comments/internal/pools"
	"github.com/booleanism/tetek/comments/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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

func (c *CommentsContr) WorkerCommentsListener(ctx context.Context) (*amqp091.Channel, error) {
	ctx, log := loggr.GetLogger(ctx, "worker")
	ch, err := c.con.Channel()
	if err != nil {
		log.Error(err, "failed to create channel")
		return nil, err
	}

	err = amqp.CommentsSetup(ch)
	if err != nil {
		log.Error(err, "failed to setup account topic")
		return nil, err
	}

	mgs, err := ch.Consume(amqp.CommentsTaskQueue, "", false, false, false, false, nil)
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

			task := amqp.CommentsTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.V(1).Info("failed to marshal account task", "requestID", d.CorrelationId, "error", err, "body", d.Body)
				res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommParseFail, Message: "error parsing comments task"})
				commentsResultPublisher(log, task, ch, d, res, err)
				continue
			}

			func() {
				cf := repo.CommentFilter{Head: task.Parent}
				fBuf, ok := pools.CommentsPool.Get().(*pools.Comments)
				if !ok {
					res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommAcquirePool, Message: "cannot acquire comments pool"})
					commentsResultPublisher(log, task, ch, d, res, fmt.Errorf("failed to acquire comments pool"))
					return
				}
				defer pools.CommentsPool.Put(fBuf)
				defer fBuf.Reset()

				if task.Cmd == 0 {
					n, err := c.repo.GetComments(ctx, cf, &fBuf.Value)
					if err == nil && n != 0 {
						res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.Success, Message: "comments found", Details: fBuf.Value[0:n]})
						commentsResultPublisher(log, task, ch, d, res, err)
						return
					}

					if err == pgx.ErrNoRows {
						res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommNoComments, Message: "comments not found"})
						commentsResultPublisher(log, task, ch, d, res, err)
						return
					}

					res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommDBError, Message: "something happen in our end"})
					commentsResultPublisher(log, task, ch, d, res, err)
					return
				}

				res, _ := json.Marshal(&amqp.CommentsResult{Code: errro.ErrCommUnknownCmd, Message: "unknown command"})
				commentsResultPublisher(log, task, ch, d, res, fmt.Errorf("unexpected command"))
			}()
		}
	}()

	return ch, nil
}

func commentsResultPublisher(log logr.Logger, task any, ch *amqp091.Channel, d amqp091.Delivery, res []byte, e error) {
	log = log.WithName("comments-result-publisher")
	err := ch.Publish(amqp.CommentsExchange, amqp.CommentsResRk, false, false, amqp091.Publishing{
		CorrelationId: d.CorrelationId,
		Body:          res,
		ContentType:   "text/json",
	})
	if err != nil {
		log.Error(err, "failed to publish comments result", "result", res, "exchange", amqp.CommentsExchange, "routing-key", amqp.CommentsResRk)
		helper.Nack(log, d, "unable to publish comments result", "task", task)
		return
	}

	if e != nil {
		helper.Nack(log, d, e.Error(), "task", task)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Error(err, "failed to ack comments task", "task", task)
		return
	}
}
