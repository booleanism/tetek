package amqp

import (
	"github.com/booleanism/tetek/account/internal/model"
	"github.com/rabbitmq/amqp091-go"
)

type User = model.User

const (
	AccountExchange     = "x_account"
	AccountTaskQueue    = "task_account"
	AccountTaskRk       = "task.account.*"
	AccountResQueue     = "res_account"
	AccountResRk        = "res.account.*"
	DlxAccountExchange  = "dlx_account"
	DlqAccountTaskQueue = "dlq_task_account"
)

type AccountTask struct {
	Cmd int `json:"cmd"`
	User
}

type AccountResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  User   `json:"account"`
}

func AccountSetup(ch *amqp091.Channel) error {
	err := ch.ExchangeDeclare(AccountExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DlxAccountExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DlqAccountTaskQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DlqAccountTaskQueue, "", DlxAccountExchange, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(AccountTaskQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DlxAccountExchange,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, AccountTaskRk, AccountExchange, false, nil)
	if err != nil {
		return err
	}

	return nil
}
