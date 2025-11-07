package amqp

import (
	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/rabbitmq/amqp091-go"
)

type User = model.User

const (
	ACCOUNT_EXCHANGE       = "x_account"
	ACCOUNT_TASK_QUEUE     = "task_account"
	ACCOUNT_TASK_RK        = "task.account.*"
	ACCOUNT_RES_QUEUE      = "res_account"
	ACCOUNT_RES_RK         = "res.account.*"
	DLX_ACCOUNT_EXCHANGE   = "dlx_account"
	DLQ_ACCOUNT_TASK_QUEUE = "dlq_task_account"
)

type AccountTask struct {
	Cmd int `json:"cmd"`
	User
}

type AccountRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  User   `json:"account"`
}

func AccountSetup(ch *amqp091.Channel) error {
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("setup account service exchange and queue")
	})
	err := ch.ExchangeDeclare(ACCOUNT_EXCHANGE, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DLX_ACCOUNT_EXCHANGE, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DLQ_ACCOUNT_TASK_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DLQ_ACCOUNT_TASK_QUEUE, "", DLX_ACCOUNT_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(ACCOUNT_TASK_QUEUE, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DLX_ACCOUNT_EXCHANGE,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, ACCOUNT_TASK_RK, ACCOUNT_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	return nil
}
