package amqp

import (
	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/rabbitmq/amqp091-go"
)

type Comment = model.Comment

const (
	CommentsExchange     = "x_comments"
	CommentsTaskQueue    = "task_comments"
	CommentsTaskRk       = "task.comments.*"
	CommentsResQueue     = "res_comments"
	CommentsResRk        = "res.comments.*"
	DlxCommentsExchange  = "dlx_comments"
	DlqCommentsTaskQueue = "dlq_task_comments"
)

type CommentsTask struct {
	Cmd int `json:"cmd"`
	Comment
}

type CommentsResult struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Details []Comment `json:"comments"`
}

func CommentsSetup(ch *amqp091.Channel) error {
	err := ch.ExchangeDeclare(CommentsExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DlxCommentsExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DlqCommentsTaskQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DlqCommentsTaskQueue, "", DlxCommentsExchange, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(CommentsTaskQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DlxCommentsExchange,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, CommentsTaskRk, CommentsExchange, false, nil)
	if err != nil {
		return err
	}

	return nil
}
