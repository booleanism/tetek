package amqp

import (
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/rabbitmq/amqp091-go"
)

type Feeds = model.Feed

const (
	FEEDS_EXCHANGE       = "x_feeds"
	FEEDS_TASK_QUEUE     = "task_feeds"
	FEEDS_TASK_RK        = "task.feeds.*"
	FEEDS_RES_QUEUE      = "res_feeds"
	FEEDS_RES_RK         = "res.feeds.*"
	DLX_FEEDS_EXCHANGE   = "dlx_feeds"
	DLQ_FEEDS_TASK_QUEUE = "dlq_task_feeds"
)

type FeedsTask struct {
	Cmd int `json:"cmd"`
	Feeds
}

type FeedsResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  Feeds  `json:"feeds"`
}

func FeedsSetup(ch *amqp091.Channel) error {
	err := ch.ExchangeDeclare(FEEDS_EXCHANGE, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DLX_FEEDS_EXCHANGE, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DLQ_FEEDS_TASK_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DLQ_FEEDS_TASK_QUEUE, "", DLX_FEEDS_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(FEEDS_TASK_QUEUE, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DLX_FEEDS_EXCHANGE,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, FEEDS_TASK_RK, FEEDS_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	return nil
}
