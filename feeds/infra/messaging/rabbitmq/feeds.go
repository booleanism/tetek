package messaging

import (
	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/rabbitmq/amqp091-go"
)

type Feeds = entities.Feed

const (
	FeedsExchange     = "x_feeds"
	FeedsTaskQueue    = "task_feeds"
	FeedsTaskRk       = "task.feeds.*"
	FeedsResQueue     = "res_feeds"
	FeedsResRk        = "res.feeds.*"
	DlxFeedsExchange  = "dlx_feeds"
	DlqFeedsTaskQueue = "dlq_task_feeds"
)

type FeedsTask struct {
	Cmd int `json:"cmd"`
	Feeds
}

type FeedsResult struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Details []Feeds `json:"feeds"`
}

func FeedsSetup(ch *amqp091.Channel) error {
	err := ch.ExchangeDeclare(FeedsExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DlxFeedsExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DlqFeedsTaskQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DlqFeedsTaskQueue, "", DlxFeedsExchange, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(FeedsTaskQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DlxFeedsExchange,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, FeedsTaskRk, FeedsExchange, false, nil)
	if err != nil {
		return err
	}

	return nil
}
